package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

func TestTroubleshootingReportIsOwnedAndExcludesPrivateInputs(t *testing.T) {
	f := setup(t)
	id := apphost.ID()
	s := Slice{ID: id, ProjectID: apphost.ID(), State: "failed", Settings: Settings{PrinterID: "private-device-id", Model: "A1 mini", NativePlate: 2, Quality: "0.20mm Standard @BBL A1M", Filaments: []string{"Generic PLA @BBL A1M"}}, Created: now()}
	_, e := f.h.DB.Exec("INSERT INTO slicer_slices VALUES(?,?,?,?,?,?,?)", s.ID, f.session.User.ID, s.ProjectID, 1, s.State, encode(s), s.Created)
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.Join(f.h.Config.Data, "slices", id)
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "engine.log"), []byte("private raw log credential marker"), 0600)
	os.WriteFile(filepath.Join(dir, "project.json"), []byte("private project name"), 0600)
	os.WriteFile(filepath.Join(dir, "result.json"), []byte(`{"return_code":-100,"plate_index":2,"sliced_plates":[{"id":2,"warning_message":"Empty layer at height 1.2"}]}`), 0600)
	w := f.call("GET", "/api/slices/"+id+"/troubleshooting", nil, nil)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	for _, secret := range []string{"private-device-id", "private raw log", "private project name", f.session.User.ID, id, dir} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatalf("report contains %s", secret)
		}
	}
	var report map[string]any
	if json.Unmarshal(w.Body.Bytes(), &report) != nil {
		t.Fatal(w.Body.String())
	}
	if report["settings"].(map[string]any)["nativePlate"] != float64(2) || !strings.Contains(w.Body.String(), "Empty layer at height 1.2") {
		t.Fatal(w.Body.String())
	}
	if f.call("GET", "/api/slices/"+apphost.ID()+"/troubleshooting", nil, nil).Code != 404 {
		t.Fatal("missing slice exposed")
	}
	_, e = f.h.DB.Exec("UPDATE slicer_slices SET user_id=? WHERE id=?", "other-owner", id)
	if e != nil {
		t.Fatal(e)
	}
	if f.call("GET", "/api/slices/"+id+"/troubleshooting", nil, nil).Code != 404 {
		t.Fatal("another user's report exposed")
	}
}
