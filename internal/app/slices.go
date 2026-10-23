package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oneworkspacestudio/my-slicer/internal/engine"
	"github.com/oneworkspacestudio/my-slicer/internal/model"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (a *App) getSlice(user, id string) (Slice, error) {
	var s Slice
	if !validID(id) {
		return s, errMissing
	}
	var raw string
	e := a.H.DB.QueryRow("SELECT data FROM slicer_slices WHERE id=? AND user_id=?", id, user).Scan(&raw)
	if e != nil {
		return s, errMissing
	}
	e = json.Unmarshal([]byte(raw), &s)
	return s, e
}
func (a *App) slices(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT data FROM slicer_slices WHERE user_id=? ORDER BY created DESC LIMIT 100", uid(r))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []Slice{}
	for rows.Next() {
		var raw string
		var s Slice
		if rows.Scan(&raw) == nil && json.Unmarshal([]byte(raw), &s) == nil {
			out = append(out, s)
		}
	}
	apphost.JSON(w, 200, out)
}
func (a *App) sliceRoute(w http.ResponseWriter, r *http.Request) {
	s, e := a.getSlice(uid(r), r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, e.Error())
		return
	}
	apphost.JSON(w, 200, s)
}

// Only inactive, unreferenced native results can be removed. IDs resolve through
// the authenticated owner and never become arbitrary filesystem paths.
func (a *App) deleteSlice(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s, e := a.getSlice(uid(r), r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, e.Error())
		return
	}
	var active int
	a.H.DB.QueryRow("SELECT COUNT(*) FROM slicer_prints WHERE slice_id=? AND state NOT IN ('completed','failed','cancelled','not_confirmed')", s.ID).Scan(&active)
	if active > 0 || s.State == "queued" || s.State == "processing" {
		apphost.Problem(w, 409, "Finish or cancel the active slice/print before removing this result")
		return
	}
	dir := filepath.Join(a.H.Config.Data, "slices", s.ID)
	if e = os.RemoveAll(dir); e != nil {
		a.problem(w, e)
		return
	}
	if _, e = a.H.DB.Exec("DELETE FROM slicer_slices WHERE id=? AND user_id=?", s.ID, uid(r)); e != nil {
		a.H.DBError(w, e)
		return
	}
	a.H.DB.Exec("DELETE FROM slicer_drive WHERE user_id=? AND kind='slice' AND object_id=?", uid(r), s.ID)
	apphost.JSON(w, 200, map[string]bool{"removed": true})
}
func (a *App) createSlice(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Revision int `json:"revision"`
	}
	if !apphost.Decode(w, r, &body) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	p, e := a.getProject(uid(r), r.PathValue("id"))
	if e != nil {
		a.problem(w, e)
		return
	}
	if p.Revision != body.Revision {
		apphost.Problem(w, 409, "This project changed; review its latest revision before slicing")
		return
	}
	if len(p.Objects) == 0 {
		apphost.Problem(w, 400, "Import a model before slicing")
		return
	}
	if e = a.validateSettings(uid(r), &p.Settings); e != nil {
		a.problem(w, e)
		return
	}
	_, profiles, ready := a.Engine.Ready()
	if !ready {
		apphost.Problem(w, 409, "The native engine is not ready; finish setup first")
		return
	}
	machine, e := profiles.Machine(p.Settings.Model)
	if e != nil {
		a.problem(w, e)
		return
	}
	if !p.Settings.PreserveImported {
		if _, e = profiles.CompatibleValues("process", p.Settings.Quality, machine.Name); e != nil {
			a.problem(w, e)
			return
		}
		if len(p.Settings.Filaments) == 0 {
			apphost.Problem(w, 400, "Select a compatible filament profile; no generic fallback will be used")
			return
		}
		for _, name := range p.Settings.Filaments {
			if _, e = profiles.CompatibleValues("filament", name, machine.Name); e != nil {
				a.problem(w, e)
				return
			}
		}
	}
	for _, o := range p.Objects {
		m, e := a.getModel(uid(r), o.ModelID)
		if e != nil {
			a.problem(w, e)
			return
		}
		if m.Kind == "stl" {
			mesh, e := a.mesh(m)
			if e != nil {
				a.problem(w, e)
				return
			}
			if e = mesh.ValidatePlate(o.Transform, machine.Polygon, machine.Exclusions, machine.Height); e != nil {
				a.problem(w, e)
				return
			}
		}
	}
	var bytes int64
	a.H.DB.QueryRow("SELECT COALESCE(SUM(json_extract(data,'$.result.bytes')),0) FROM slicer_slices WHERE user_id=?", uid(r)).Scan(&bytes)
	if bytes > 4<<30 {
		apphost.Problem(w, 413, "Retained sliced archives exceed 4 GiB. Remove old results before slicing again")
		return
	}
	var queued, count int
	a.H.DB.QueryRow("SELECT COUNT(*) FROM slicer_slices WHERE user_id=? AND state IN ('queued','processing')", uid(r)).Scan(&queued)
	a.H.DB.QueryRow("SELECT COUNT(*) FROM slicer_slices WHERE user_id=?", uid(r)).Scan(&count)
	if queued >= 5 || count >= 50 {
		apphost.Problem(w, 429, "Limit of5 active slices or50 retained results reached; remove an old result before adding another")
		return
	}
	s := Slice{ID: apphost.ID(), ProjectID: p.ID, Revision: p.Revision, State: "queued", Message: "Waiting for the single native slice worker", Settings: p.Settings, Created: now(), Updated: now()}
	dir := filepath.Join(a.H.Config.Data, "slices", s.ID)
	if e = os.Mkdir(dir, 0700); e != nil {
		a.problem(w, e)
		return
	}
	if e = engine.WriteJSON(filepath.Join(dir, "project.json"), p); e != nil {
		a.problem(w, e)
		return
	}
	if _, e = a.H.DB.Exec("INSERT INTO slicer_slices VALUES(?,?,?,?,?,?,?)", s.ID, uid(r), p.ID, p.Revision, s.State, encode(s), s.Created); e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 202, s)
}
func (a *App) storeSlice(user string, s Slice) {
	s.Updated = now()
	a.H.DB.Exec("UPDATE slicer_slices SET state=?,data=? WHERE id=? AND user_id=?", s.State, encode(s), s.ID, user)
}
func (a *App) runNextSlice(ctx context.Context) {
	a.mu.Lock()
	var user, raw string
	e := a.H.DB.QueryRow("SELECT user_id,data FROM slicer_slices WHERE state='queued' ORDER BY created LIMIT 1").Scan(&user, &raw)
	if e != nil {
		a.mu.Unlock()
		return
	}
	var s Slice
	if json.Unmarshal([]byte(raw), &s) != nil {
		a.mu.Unlock()
		return
	}
	root, profiles, ready := a.Engine.Ready()
	if !ready {
		a.mu.Unlock()
		return
	}
	s.State = "processing"
	s.Message = "Resolving immutable project revision and official profiles"
	a.storeSlice(user, s)
	worker, cancel := context.WithTimeout(ctx, 25*time.Minute)
	a.cancelSlice[s.ID] = cancel
	a.mu.Unlock()
	defer func() { cancel(); a.mu.Lock(); delete(a.cancelSlice, s.ID); a.mu.Unlock() }()
	a.H.Freeze.RLock()
	defer a.H.Freeze.RUnlock()
	dir := filepath.Join(a.H.Config.Data, "slices", s.ID)
	fail := func(err error) {
		s.State = "failed"
		s.Message = err.Error()
		if worker.Err() != nil {
			s.State = "cancelled"
			s.Message = "Slicing cancelled or time limit reached; originals are preserved"
		}
		s.Message = strings.ReplaceAll(s.Message, a.H.Config.Data, "[app data]")
		a.storeSlice(user, s)
	}
	var p Project
	b, e := os.ReadFile(filepath.Join(dir, "project.json"))
	if e != nil || json.Unmarshal(b, &p) != nil {
		fail(errors.New("The immutable project snapshot is unavailable"))
		return
	}
	machine, e := profiles.Machine(s.Settings.Model)
	if e != nil {
		fail(e)
		return
	}
	config := engine.SliceConfig{Machine: machine.Values, Plate: s.Settings.NativePlate, Overrides: engine.Values{"enable_arc_fitting": "0", "curr_bed_type": s.Settings.Plate, "print_sequence": "by layer"}, PreserveImported: s.Settings.PreserveImported}
	if !s.Settings.PreserveImported {
		config.Process, e = profiles.CompatibleValues("process", s.Settings.Quality, machine.Name)
		if e != nil {
			fail(e)
			return
		}
		for _, name := range s.Settings.Filaments {
			v, e := profiles.CompatibleValues("filament", name, machine.Name)
			if e != nil {
				fail(e)
				return
			}
			config.Filaments = append(config.Filaments, v)
		}
	}
	override := func(key string, value any) {
		if !s.Settings.PreserveImported || s.Settings.Overrides[key] {
			config.Overrides[key] = value
		}
	}
	override("layer_height", strconv.FormatFloat(s.Settings.LayerHeight, 'f', -1, 64))
	override("sparse_infill_density", fmt.Sprintf("%d%%", s.Settings.Infill))
	override("wall_loops", strconv.Itoa(s.Settings.Walls))
	support := "0"
	if s.Settings.Supports {
		support = "1"
	}
	override("enable_support", support)
	override("brim_type", s.Settings.Brim)
	for key, value := range s.Settings.Process {
		override(key, value)
	}
	var objects []model.Object
	for _, instance := range p.Objects {
		m, e := a.getModel(user, instance.ModelID)
		if e != nil {
			fail(e)
			return
		}
		if m.Kind == "3mf" {
			config.Input = filepath.Join(a.H.Config.Data, "models", m.ID, "original.3mf")
			if _, e = model.InspectProject(config.Input); e != nil {
				fail(e)
				return
			}
			break
		}
		mesh, e := a.mesh(m)
		if e != nil {
			fail(e)
			return
		}
		objects = append(objects, model.Object{Mesh: mesh, Transform: instance.Transform})
	}
	if config.Input == "" {
		config.Plate = 1
		config.Input = filepath.Join(dir, "input.3mf")
		if e = model.WriteProject(config.Input, objects); e != nil {
			fail(e)
			return
		}
	}
	s.Message = "Native Bambu Studio is slicing the selected plate. Stage progress is genuine; elapsed time continues while the engine runs."
	a.storeSlice(user, s)
	result, e := engine.Slice(worker, root, dir, config)
	if e != nil {
		fail(e)
		return
	}
	want := "N2S"
	if s.Settings.Model == "A1 mini" {
		want = "N1"
	}
	if result.ModelID != want || result.PrinterName != machine.Name || result.PlateType != s.Settings.Plate {
		fail(errors.New("Native output target/nozzle/plate does not match the immutable selected configuration"))
		return
	}
	s.State = "ready"
	s.Message = "Native slice complete. Review the actual layer toolpath and selected printer before explicitly choosing Print."
	s.Result = &result
	a.storeSlice(user, s)
	a.queueDrive(user, "slice", s.ID, cleanName(p.Name)+"-r"+strconv.Itoa(s.Revision)+".gcode.3mf")
}
func (a *App) cancelSliceRoute(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s, e := a.getSlice(uid(r), r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, e.Error())
		return
	}
	if cancel := a.cancelSlice[s.ID]; cancel != nil {
		cancel()
	} else if s.State == "queued" {
		s.State = "cancelled"
		s.Message = "Cancelled before slicing"
		a.storeSlice(uid(r), s)
	}
	apphost.JSON(w, 202, s)
}
func (a *App) preview(w http.ResponseWriter, r *http.Request) {
	s, e := a.getSlice(uid(r), r.PathValue("id"))
	if e != nil || s.State != "ready" || s.Result == nil {
		apphost.Problem(w, 404, "A completed native slice is required for toolpath preview")
		return
	}
	from, _ := strconv.Atoi(r.URL.Query().Get("from"))
	to, _ := strconv.Atoi(r.URL.Query().Get("to"))
	if from < 1 {
		from = 1
	}
	if to < from {
		to = from
	}
	if to > s.Result.Layers || to-from > 49 {
		apphost.Problem(w, 413, "Choose at most50 valid layers at a time for this browser preview")
		return
	}
	layers := []model.Layer{}
	total := 0
	for n := from; n <= to; n++ {
		raw, e := os.ReadFile(filepath.Join(a.H.Config.Data, "slices", s.ID, "preview", fmt.Sprintf("layer-%d.json", n)))
		if e != nil {
			apphost.Problem(w, 404, "This native layer preview is unavailable")
			return
		}
		var layer model.Layer
		if json.Unmarshal(raw, &layer) != nil {
			apphost.Problem(w, 500, "Layer data could not be read")
			return
		}
		total += len(layer.Segments)
		if total > 250000 {
			apphost.Problem(w, 413, "This selection has over250,000 segments. Narrow the layer range; the complete native archive remains unchanged")
			return
		}
		layers = append(layers, layer)
	}
	apphost.JSON(w, 200, map[string]any{"layers": layers, "features": s.Result.Features, "segments": total, "source": "native-gcode"})
}
func (a *App) sliceArchive(w http.ResponseWriter, r *http.Request) {
	s, e := a.getSlice(uid(r), r.PathValue("id"))
	if e != nil || s.State != "ready" {
		apphost.Problem(w, 404, "The completed sliced archive is unavailable")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="my-slicer.gcode.3mf"`)
	http.ServeFile(w, r, filepath.Join(a.H.Config.Data, "slices", s.ID, "result.gcode.3mf"))
}
