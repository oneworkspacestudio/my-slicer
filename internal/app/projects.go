package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oneworkspacestudio/my-slicer/internal/engine"
	"github.com/oneworkspacestudio/my-slicer/internal/model"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Object struct {
	ID        string          `json:"id"`
	ModelID   string          `json:"modelID"`
	Transform model.Transform `json:"transform"`
}
type Model struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Kind             string            `json:"kind"`
	Size             int64             `json:"size"`
	SHA256           string            `json:"sha256"`
	Min              model.Vec         `json:"min"`
	Max              model.Vec         `json:"max"`
	Triangles        int               `json:"triangles"`
	PreviewTriangles int               `json:"previewTriangles"`
	Import           *model.ImportInfo `json:"import"`
	Created          string            `json:"created"`
}

func defaults() Settings {
	return Settings{Model: "A1", Nozzle: "0.4", Plate: "Textured PEI Plate", Quality: "0.20mm Standard @BBL A1", Filaments: []string{"Bambu PLA Basic @BBL A1"}, NativePlate: 1, LayerHeight: .2, Infill: 15, Walls: 2, Brim: "auto_brim", Overrides: map[string]bool{}}
}
func (a *App) getProject(user, id string) (Project, error) {
	var p Project
	if !validID(id) {
		return p, errMissing
	}
	var raw string
	e := a.H.DB.QueryRow("SELECT data FROM slicer_projects WHERE id=? AND user_id=?", id, user).Scan(&raw)
	if e != nil {
		return p, errMissing
	}
	e = json.Unmarshal([]byte(raw), &p)
	return p, e
}
func (a *App) getModel(user, id string) (Model, error) {
	var m Model
	if !validID(id) {
		return m, errMissing
	}
	var raw string
	e := a.H.DB.QueryRow("SELECT data FROM slicer_models WHERE id=? AND user_id=?", id, user).Scan(&raw)
	if e != nil {
		return m, errMissing
	}
	e = json.Unmarshal([]byte(raw), &m)
	return m, e
}
func (a *App) projects(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT data FROM slicer_projects WHERE user_id=? ORDER BY updated DESC LIMIT 100", uid(r))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []Project{}
	for rows.Next() {
		var raw string
		var p Project
		if e = rows.Scan(&raw); e != nil {
			a.H.DBError(w, e)
			return
		}
		if json.Unmarshal([]byte(raw), &p) == nil {
			out = append(out, p)
		}
	}
	apphost.JSON(w, 200, out)
}
func (a *App) projectRoute(w http.ResponseWriter, r *http.Request) {
	p, e := a.getProject(uid(r), r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, e.Error())
		return
	}
	apphost.JSON(w, 200, p)
}
func (a *App) createProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name  string `json:"name"`
		Draft bool   `json:"draft"`
	}
	if !apphost.Decode(w, r, &body) {
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" || len(body.Name) > 160 {
		apphost.Problem(w, 400, "Name your project using1–160 characters")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var count int
	a.H.DB.QueryRow("SELECT COUNT(*) FROM slicer_projects WHERE user_id=?", uid(r)).Scan(&count)
	if count >= 100 {
		apphost.Problem(w, 413, "This account has 100 saved projects and recent plates. Open an existing plate to keep working.")
		return
	}
	p := Project{ID: apphost.ID(), Name: body.Name, Draft: body.Draft, Revision: 1, Settings: defaults(), Objects: []Object{}, Updated: now()}
	dir := filepath.Join(a.H.Config.Data, "projects", p.ID)
	if e := os.Mkdir(dir, 0700); e != nil {
		a.problem(w, e)
		return
	}
	if e := engine.WriteJSON(filepath.Join(dir, "revision-1.json"), p); e != nil {
		a.problem(w, e)
		return
	}
	if _, e := a.H.DB.Exec("INSERT INTO slicer_projects VALUES(?,?,?,?,?)", p.ID, uid(r), p.Revision, encode(p), p.Updated); e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 201, p)
}
func (a *App) validateSettings(user string, s *Settings) error {
	if e := engine.ValidateProcess(s.Process); e != nil {
		return e
	}
	if s.Model != "A1" && s.Model != "A1 mini" || s.Nozzle != "0.4" {
		return errors.New("Select a supported A1/A1 mini0.4mm profile")
	}
	if s.Plate != "Textured PEI Plate" && s.Plate != "High Temp Plate" {
		return errors.New("Choose the physically installed supported PEI plate")
	}
	if s.NativePlate < 1 || s.NativePlate > 100 || len(s.Filaments) > 16 {
		return errors.New("Check plate and filament selection")
	}
	if !model.ValidNumber(s.LayerHeight) || s.LayerHeight < .08 || s.LayerHeight > .28 || s.Infill < 0 || s.Infill > 100 || s.Walls < 1 || s.Walls > 20 {
		return errors.New("Layer height must be0.08–0.28mm, infill0–100% and walls1–20")
	}
	switch s.Brim {
	case "auto_brim", "no_brim", "outer_only", "inner_only", "outer_and_inner":
	default:
		return errors.New("Choose a supported brim option")
	}
	if s.PrinterID != "" {
		printer, _, e := a.getPrinter(user, s.PrinterID)
		if e != nil {
			return e
		}
		if s.Model != printer.Model || s.Nozzle != printer.Nozzle || s.Plate != printer.Plate {
			return errors.New("Project target, nozzle and plate must match the saved printer configuration")
		}
		s.PrinterRevision = printer.Revision
	} else {
		s.PrinterRevision = 0
	}
	if s.Overrides == nil {
		s.Overrides = map[string]bool{}
	}
	return nil
}
func (a *App) saveProject(user string, p Project, previous int) (Project, error) {
	if e := a.validateSettings(user, &p.Settings); e != nil {
		return p, e
	}
	if strings.TrimSpace(p.Name) == "" || len(p.Name) > 160 || len(p.Objects) > 100 {
		return p, errors.New("Check project name and the100 object-instance limit")
	}
	seen := map[string]bool{}
	var native *Model
	for _, obj := range p.Objects {
		if !validID(obj.ID) || seen[obj.ID] || !obj.Transform.Valid() {
			return p, errors.New("Check unique object IDs, positions, rotations and uniform scale")
		}
		seen[obj.ID] = true
		m, e := a.getModel(user, obj.ModelID)
		if e != nil {
			return p, e
		}
		if m.Kind == "3mf" {
			copy := m
			native = &copy
		}
	}
	if native != nil {
		if len(p.Objects) != 1 {
			return p, errors.New("A preserved native3MF project must use its own project; mixing or duplicating it would discard project data")
		}
		if p.Objects[0].Transform != (model.Transform{Scale: 1}) {
			return p, errors.New("Browser transforms apply to STL objects. Native3MF placement/painting/modifiers remain intact; edit them in the source project or re-import an explicitly geometry-only STL")
		}
		found := false
		for _, plate := range native.Import.Plates {
			if plate.Index == p.Settings.NativePlate {
				found = true
			}
		}
		if !found {
			return p, errors.New("Select an existing plate from the native project")
		}
		if !p.Settings.PreserveImported && len(native.Import.Filaments) > 0 && len(p.Settings.Filaments) != len(native.Import.Filaments) {
			return p, fmt.Errorf("Choose a material for each of this project's %d filament slots in Prepare", len(native.Import.Filaments))
		}
		if p.Settings.PreserveImported {
			if native.Import.Printer != "Bambu Lab "+p.Settings.Model || len(native.Import.Nozzles) != 1 || native.Import.Nozzles[0] != "0.4" {
				return p, errors.New("Changing native printer/nozzle requires a fresh compatible quality and filament selection; turn off Preserve imported settings before re-slicing")
			}
		}
	} else if p.Settings.PreserveImported {
		return p, errors.New("Only a complete imported native3MF can preserve embedded project settings")
	}
	p.Revision = previous + 1
	p.Updated = now()
	path := filepath.Join(a.H.Config.Data, "projects", p.ID, fmt.Sprintf("revision-%d.json", p.Revision))
	if e := engine.WriteJSON(path, p); e != nil {
		return p, e
	}
	result, e := a.H.DB.Exec("UPDATE slicer_projects SET revision=?,data=?,updated=? WHERE id=? AND user_id=? AND revision=?", p.Revision, encode(p), p.Updated, p.ID, user, previous)
	if e != nil {
		return p, e
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return p, errors.New("This project changed in another tab; reload before applying edits")
	}
	return p, nil
}
func (a *App) updateProject(w http.ResponseWriter, r *http.Request) {
	var body Project
	if !apphost.Decode(w, r, &body) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	old, e := a.getProject(uid(r), r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, e.Error())
		return
	}
	if old.Revision != body.Revision {
		apphost.Problem(w, 409, "This project changed in another tab; reload to keep the latest edits")
		return
	}
	body.ID = old.ID
	p, e := a.saveProject(uid(r), body, old.Revision)
	if e != nil {
		a.problem(w, e)
		return
	}
	apphost.JSON(w, 200, p)
}
func (a *App) models(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT data FROM slicer_models WHERE user_id=? ORDER BY created DESC LIMIT 500", uid(r))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []Model{}
	for rows.Next() {
		var raw string
		var m Model
		if e = rows.Scan(&raw); e != nil {
			a.H.DBError(w, e)
			return
		}
		if json.Unmarshal([]byte(raw), &m) == nil {
			out = append(out, m)
		}
	}
	apphost.JSON(w, 200, out)
}
func (a *App) uploadModel(w http.ResponseWriter, r *http.Request) {
	a.storageMu.Lock()
	defer a.storageMu.Unlock()
	a.mu.Lock()
	defer a.mu.Unlock()
	p, e := a.getProject(uid(r), r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, e.Error())
		return
	}
	revision, _ := strconv.Atoi(r.Header.Get("X-Project-Revision"))
	if revision != p.Revision {
		apphost.Problem(w, 409, "Reload the project before uploading into a newer revision")
		return
	}
	name := strings.TrimSpace(r.Header.Get("X-File-Name"))
	kind := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
	if name == "" || len(name) > 200 || strings.ContainsAny(name, "/\\\r\n\x00") || (kind != "stl" && kind != "3mf") {
		apphost.Problem(w, 400, "Choose an STL or self-contained Bambu3MF with a short filename")
		return
	}
	if len(p.Objects) >= 100 || kind == "3mf" && len(p.Objects) > 0 {
		apphost.Problem(w, 409, "Import native3MF into a new project so all project metadata is preserved")
		return
	}
	for _, o := range p.Objects {
		m, e := a.getModel(uid(r), o.ModelID)
		if e != nil || m.Kind == "3mf" {
			apphost.Problem(w, 409, "This project contains preserved3MF data. Start a new STL project to add separate objects")
			return
		}
	}
	var modelCount int
	a.H.DB.QueryRow("SELECT COUNT(*) FROM slicer_models WHERE user_id=?", uid(r)).Scan(&modelCount)
	if modelCount >= 500 {
		apphost.Problem(w, 413, "Limit of 500 retained models reached")
		return
	}
	var used int64
	a.H.DB.QueryRow("SELECT COALESCE(SUM(bytes),0) FROM slicer_models WHERE user_id=?", uid(r)).Scan(&used)
	if used >= 2<<30 {
		apphost.Problem(w, 413, "This account's model storage budget is2GiB")
		return
	}
	limit := min(int64(100<<20), (2<<30)-used)
	_ = http.NewResponseController(w).SetReadDeadline(time.Now().Add(3 * time.Minute))
	defer http.NewResponseController(w).SetReadDeadline(time.Time{})
	data, e := io.ReadAll(http.MaxBytesReader(w, r.Body, limit))
	if e != nil || len(data) == 0 {
		apphost.Problem(w, 413, "Upload failed or exceeds100MiB")
		return
	}
	sum := sha256.Sum256(data)
	m := Model{ID: apphost.ID(), Name: name, Kind: kind, Size: int64(len(data)), SHA256: hex.EncodeToString(sum[:]), Created: now()}
	dir := filepath.Join(a.H.Config.Data, "models", m.ID)
	if e = os.Mkdir(dir, 0700); e != nil {
		a.problem(w, e)
		return
	}
	kept := false
	defer func() {
		if !kept {
			_ = os.RemoveAll(dir)
		}
	}()
	original := filepath.Join(dir, "original."+kind)
	if e = os.WriteFile(original, data, 0600); e != nil {
		a.problem(w, e)
		return
	}
	transform := model.Transform{Scale: 1}
	if kind == "stl" {
		mesh, e := model.ParseSTL(data)
		if e != nil {
			a.problem(w, e)
			return
		}
		mesh = mesh.Centred()
		m.Min = mesh.Min
		m.Max = mesh.Max
		m.Triangles = len(mesh.Triangles)
		stride := max(1, (m.Triangles+99999)/100000)
		m.PreviewTriangles = (m.Triangles + stride - 1) / stride
		if e = os.WriteFile(filepath.Join(dir, "mesh.stl"), mesh.Binary(0), 0600); e != nil {
			a.problem(w, e)
			return
		}
		if e = os.WriteFile(filepath.Join(dir, "preview.stl"), mesh.Binary(100000), 0600); e != nil {
			a.problem(w, e)
			return
		}
		width := 256.0
		if p.Settings.Model == "A1 mini" {
			width = 180
		}
		transform.X = width / 2
		transform.Y = width / 2
	} else {
		info, e := model.InspectProject(original)
		if e != nil {
			a.problem(w, e)
			return
		}
		m.Import = &info
		p.Settings.NativePlate = info.Plates[0].Index
		p.Settings.PreserveImported = info.PreserveSettings
		if info.PreserveSettings {
			switch info.Printer {
			case "Bambu Lab A1":
				p.Settings.Model = "A1"
			case "Bambu Lab A1 mini":
				p.Settings.Model = "A1 mini"
			default:
				p.Settings.PreserveImported = false
				info.Notice = fmt.Sprintf("This project was made for %s. Choose your A1 or A1 mini and a matching material for each filament below before slicing. The original file, colours and model layout are kept.", info.Printer)
			}
			p.Settings.Filaments = info.Filaments
			if !p.Settings.PreserveImported {
				// Retargeting is explicit, slot for slot; do not silently map a
				// multi-material project onto the previous single-filament choice.
				p.Settings.Filaments = make([]string, len(info.Filaments))
			}
			if v, ok := info.Settings["layer_height"].(string); ok {
				if n, e := strconv.ParseFloat(v, 64); e == nil {
					p.Settings.LayerHeight = n
				}
			}
			if v, ok := info.Settings["sparse_infill_density"].(string); ok {
				if n, e := strconv.Atoi(strings.TrimSuffix(v, "%")); e == nil {
					p.Settings.Infill = n
				}
			}
			p.Settings.PrinterID = ""
		}
	}
	if _, e = a.H.DB.Exec("INSERT INTO slicer_models VALUES(?,?,?,?,?)", m.ID, uid(r), encode(m), m.Size, m.Created); e != nil {
		a.H.DBError(w, e)
		return
	}
	kept = true
	p.Objects = append(p.Objects, Object{ID: apphost.ID(), ModelID: m.ID, Transform: transform})
	p, e = a.saveProject(uid(r), p, p.Revision)
	if e != nil {
		a.problem(w, e)
		return
	}
	if r.Header.Get("X-From-Drive") != "true" {
		a.queueDrive(uid(r), "model", m.ID, m.Name)
	}
	apphost.JSON(w, 201, map[string]any{"model": m, "project": p})
}
func (a *App) modelFile(w http.ResponseWriter, r *http.Request) {
	m, e := a.getModel(uid(r), r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, e.Error())
		return
	}
	file := "original." + m.Kind
	if strings.HasSuffix(r.URL.Path, "/mesh") {
		if m.Kind != "stl" {
			plate, err := strconv.Atoi(r.URL.Query().Get("plate"))
			if err != nil || plate < 1 || plate > 100 {
				apphost.Problem(w, 400, "Choose an imported plate to view")
				return
			}
			file = fmt.Sprintf("preview-plate-%d.stl", plate)
			cached := filepath.Join(a.H.Config.Data, "models", m.ID, file)
			a.storageMu.Lock()
			if _, err = os.Stat(cached); os.IsNotExist(err) {
				var mesh model.Mesh
				mesh, err = model.PreviewProject(filepath.Join(a.H.Config.Data, "models", m.ID, "original.3mf"), plate)
				if err == nil {
					err = os.WriteFile(cached+".tmp", mesh.Binary(100000), 0600)
					if err == nil {
						err = os.Rename(cached+".tmp", cached)
					}
				}
			}
			a.storageMu.Unlock()
			if err != nil {
				apphost.Problem(w, 422, "Could not display this plate. "+err.Error())
				return
			}
		}
		if m.Kind == "stl" {
			file = "preview.stl"
		}
		w.Header().Set("Content-Type", "model/stl")
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="model.`+m.Kind+`"`)
	}
	http.ServeFile(w, r, filepath.Join(a.H.Config.Data, "models", m.ID, file))
}
func (a *App) mesh(m Model) (model.Mesh, error) {
	b, e := os.ReadFile(filepath.Join(a.H.Config.Data, "models", m.ID, "mesh.stl"))
	if e != nil {
		return model.Mesh{}, e
	}
	return model.ParseSTL(b)
}
func (a *App) arrange(w http.ResponseWriter, r *http.Request) {
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
	if body.Revision != p.Revision {
		apphost.Problem(w, 409, "Reload this project's latest revision")
		return
	}
	_, profiles, ready := a.Engine.Ready()
	if !ready {
		apphost.Problem(w, 409, "Install the engine before arranging against its precise machine geometry")
		return
	}
	machine, e := profiles.Machine(p.Settings.Model)
	if e != nil {
		a.problem(w, e)
		return
	}
	x, y, row := 3.0, 3.0, 0.0
	for i, obj := range p.Objects {
		m, e := a.getModel(uid(r), obj.ModelID)
		if e != nil || m.Kind != "stl" {
			apphost.Problem(w, 409, "Native3MF plate layout is preserved; browser arrange is available for STL projects")
			return
		}
		mesh, e := a.mesh(m)
		if e != nil {
			a.problem(w, e)
			return
		}
		t := obj.Transform
		t.X = 0
		t.Y = 0
		t.Z = 0
		var triangles []model.Triangle
		for _, tri := range mesh.Triangles {
			var transformed model.Triangle
			for n, v := range tri {
				transformed[n] = t.Apply(v)
			}
			triangles = append(triangles, transformed)
		}
		lo, hi, e := model.Bounds(triangles)
		if e != nil {
			a.problem(w, e)
			return
		}
		width, depth := hi[0]-lo[0], hi[1]-lo[1]
		if x+width > machine.Width-3 {
			x = 3
			y += row + 3
			row = 0
		}
		t.X = x - lo[0]
		t.Y = y - lo[1]
		t.Z = -lo[2]
		if e = mesh.ValidatePlate(t, machine.Polygon, machine.Exclusions, machine.Height); e != nil {
			apphost.Problem(w, 400, "Objects do not all fit with3mm spacing. Rotate, scale or remove an object before arranging")
			return
		}
		p.Objects[i].Transform = t
		x += width + 3
		row = max(row, depth)
	}
	p, e = a.saveProject(uid(r), p, p.Revision)
	if e != nil {
		a.problem(w, e)
		return
	}
	apphost.JSON(w, 200, p)
}
func cleanName(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < ' ' || strings.ContainsRune(`/\:*?"<>|`, r) {
			return '_'
		}
		return r
	}, s)
	if len(s) > 120 {
		s = s[:120]
	}
	return strings.TrimSpace(s)
}
func copyBytes(path string, b []byte) error { return os.WriteFile(path, bytes.Clone(b), 0600) }
