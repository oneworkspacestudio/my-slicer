package app

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/oneworkspacestudio/my-slicer/internal/engine"
	"github.com/oneworkspacestudio/my-slicer/internal/model"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fixture struct {
	a       *App
	h       *apphost.Host
	cookie  *http.Cookie
	session apphost.Session
	t       *testing.T
}

func setup(t *testing.T) *fixture {
	h, e := apphost.New(apphost.Config{ID: "slicer", Name: "My Slicer", Version: "test", Data: filepath.Join(t.TempDir(), "data"), Listen: "127.0.0.1:29020", Origin: "http://localhost:29020"}, nil)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { h.Close() })
	a, e := New(h)
	if e != nil {
		t.Fatal(e)
	}
	f := &fixture{a: a, h: h, t: t}
	if e = h.Bootstrap("owner@example.test", "fixture-password"); e != nil {
		t.Fatal(e)
	}
	w := f.call("POST", "/api/login", map[string]string{"email": "owner@example.test", "password": "fixture-password"}, nil)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	f.cookie = w.Result().Cookies()[0]
	json.Unmarshal(f.call("GET", "/api/me", nil, nil).Body.Bytes(), &f.session)
	return f
}
func (f *fixture) call(method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	var reader io.Reader
	if b, ok := body.([]byte); ok {
		reader = bytes.NewReader(b)
	} else if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	r := httptest.NewRequest(method, f.h.Config.Origin+path, reader)
	r.RemoteAddr = "127.0.0.1:12345"
	r.Header.Set("Origin", f.h.Config.Origin)
	r.Header.Set("X-CSRF-Token", f.session.CSRF)
	if f.cookie != nil {
		r.AddCookie(f.cookie)
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	f.h.Handler().ServeHTTP(w, r)
	return w
}
func (f *fixture) project() Project {
	w := f.call("POST", "/api/projects", map[string]string{"name": "Native test cube"}, nil)
	if w.Code != 201 {
		f.t.Fatal(w.Code, w.Body.String())
	}
	var p Project
	json.Unmarshal(w.Body.Bytes(), &p)
	return p
}
func (f *fixture) upload(p Project, data []byte, name string) Project {
	w := f.call("POST", "/api/projects/"+p.ID+"/models", data, map[string]string{"X-File-Name": name, "X-Project-Revision": stringInt(p.Revision)})
	if w.Code != 201 {
		f.t.Fatal(w.Code, w.Body.String())
	}
	var result struct{ Project Project }
	json.Unmarshal(w.Body.Bytes(), &result)
	return result.Project
}
func stringInt(n int) string { b, _ := json.Marshal(n); return string(b) }
func TestProjectOwnershipRevisionAndInvalidUpload(t *testing.T) {
	f := setup(t)
	p := f.project()
	p = f.upload(p, model.Cube(10).Binary(0), "cube.stl")
	if p.Revision != 2 || len(p.Objects) != 1 {
		t.Fatal(p)
	}
	w := f.call("PATCH", "/api/projects/"+p.ID, p, nil)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = f.call("PATCH", "/api/projects/"+p.ID, p, nil); w.Code != 409 {
		t.Fatal("stale revision accepted", w.Code)
	}
	dirs, _ := os.ReadDir(filepath.Join(f.h.Config.Data, "models"))
	w = f.call("POST", "/api/projects/"+p.ID+"/models", []byte("invalid STL"), map[string]string{"X-File-Name": "bad.stl", "X-Project-Revision": "3"})
	if w.Code == 201 {
		t.Fatal("invalid STL accepted")
	}
	after, _ := os.ReadDir(filepath.Join(f.h.Config.Data, "models"))
	if len(after) != len(dirs) {
		t.Fatal("invalid upload leaked files")
	}
	if w = f.call("GET", "/api/projects/"+p.ID, nil, map[string]string{"X-Workspace-Expected-Account": "other-user"}); w.Code != 409 {
		t.Fatal("account mismatch accepted")
	}
	if w = f.call("POST", "/api/local-users", map[string]string{"email": "member@example.test", "password": "fixture-password"}, nil); w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	login := f.call("POST", "/api/login", map[string]string{"email": "member@example.test", "password": "fixture-password"}, nil)
	f.cookie = login.Result().Cookies()[0]
	json.Unmarshal(f.call("GET", "/api/me", nil, nil).Body.Bytes(), &f.session)
	if w = f.call("GET", "/api/projects/"+p.ID, nil, nil); w.Code != 404 {
		t.Fatal("cross-account project exposed")
	}
	if w = f.call("POST", "/api/setup/engine", nil, nil); w.Code != 403 {
		t.Fatal("member installed engine")
	}
}
func TestDurablePrintMerge(t *testing.T) {
	f := setup(t)
	p := Print{ID: apphost.ID(), State: "printing", Done: 50, CancelRequested: true}
	_, e := f.h.DB.Exec("INSERT INTO slicer_prints VALUES(?,?,?,?,?,?,?,?,?,?)", p.ID, f.session.User.ID, "printer", "slice", "physical", "confirmation", "hash", p.State, encode(p), now())
	if e != nil {
		t.Fatal(e)
	}
	stale := p
	stale.State = "awaiting_printer"
	stale.CancelRequested = false
	stale.Done = 1
	if e = f.a.storePrint(f.session.User.ID, stale); e != nil {
		t.Fatal(e)
	}
	got, _ := f.a.getPrint(f.session.User.ID, p.ID)
	if got.State != "printing" || !got.CancelRequested || got.Done != 50 {
		t.Fatal(got)
	}
	got.State = "completed"
	f.a.storePrint(f.session.User.ID, got)
	f.a.storePrint(f.session.User.ID, stale)
	got, _ = f.a.getPrint(f.session.User.ID, p.ID)
	if got.State != "completed" {
		t.Fatal("terminal state overwritten")
	}
}
func TestRealNativeHTTPWorkflow(t *testing.T) {
	archive := os.Getenv("MYSLICER_TEST_ENGINE_ARCHIVE")
	if archive == "" {
		t.Skip("Set reviewed official native archive for real Windows acceptance")
	}
	t.Setenv("MYSLICER_ENGINE_ARCHIVE", archive)
	f := setup(t)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	f.a.Engine.Install(ctx)
	for f.a.Engine.State().Busy && ctx.Err() == nil {
		time.Sleep(100 * time.Millisecond)
	}
	defer f.a.Engine.Wait()
	if f.a.Engine.State().Phase != "ready" {
		t.Fatal(f.a.Engine.State())
	}
	for _, name := range []string{"A1", "A1 mini"} {
		p := f.project()
		p = f.upload(p, model.Cube(10).Binary(0), "cube.stl")
		_, profiles, _ := f.a.Engine.Ready()
		m, e := profiles.Machine(name)
		if e != nil {
			t.Fatal(e)
		}
		p.Settings.Model = name
		p.Settings.Quality = engine.Text(m.Values["default_print_profile"])
		p.Settings.Filaments = []string{engine.Text(m.Values["default_filament_profile"])}
		w := f.call("PATCH", "/api/projects/"+p.ID, p, nil)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		json.Unmarshal(w.Body.Bytes(), &p)
		w = f.call("POST", "/api/projects/"+p.ID+"/slices", map[string]int{"revision": p.Revision}, nil)
		if w.Code != 202 {
			t.Fatal(w.Code, w.Body.String())
		}
		var s Slice
		json.Unmarshal(w.Body.Bytes(), &s)
		f.a.runNextSlice(ctx)
		s, _ = f.a.getSlice(f.session.User.ID, s.ID)
		if s.State != "ready" || s.Result == nil {
			t.Fatal(name, s.State, s.Message)
		}
		if s.Result.PrinterName != m.Name || s.Result.PlateType != p.Settings.Plate || s.Result.PreviewSegments == 0 {
			t.Fatal(s.Result)
		}
		if w = f.call("GET", "/api/slices/"+s.ID+"/preview?from=1", nil, nil); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		archivePath := filepath.Join(f.h.Config.Data, "slices", s.ID, "result.gcode.3mf")
		data, e := os.ReadFile(archivePath)
		if e != nil {
			t.Fatal(e)
		}
		round := f.project()
		round = f.upload(round, data, "roundtrip.3mf")
		if !round.Settings.PreserveImported {
			t.Fatal("native settings lost")
		}
		w = f.call("POST", "/api/projects/"+round.ID+"/slices", map[string]int{"revision": round.Revision}, nil)
		if w.Code != 202 {
			t.Fatal(w.Code, w.Body.String())
		}
		var roundSlice Slice
		json.Unmarshal(w.Body.Bytes(), &roundSlice)
		f.a.runNextSlice(ctx)
		roundSlice, _ = f.a.getSlice(f.session.User.ID, roundSlice.ID)
		if roundSlice.State != "ready" {
			t.Fatal("native roundtrip", name, roundSlice.Message)
		}
		if w = f.call("DELETE", "/api/slices/"+s.ID, nil, nil); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
