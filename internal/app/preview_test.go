package app

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/oneworkspacestudio/my-slicer/internal/model"
	"os"
	"path/filepath"
	"testing"
)

func TestNativeMeshPreviewAndOriginalRemainIndependent(t *testing.T) {
	f := setup(t)
	file := filepath.Join(t.TempDir(), "native.3mf")
	if e := model.WriteProject(file, []model.Object{{Mesh: model.Cube(10), Transform: model.Transform{Scale: 1, X: 90, Y: 90}}}); e != nil {
		t.Fatal(e)
	}
	data, _ := os.ReadFile(file)
	p := f.upload(f.project(), data, "native.3mf")
	id := p.Objects[0].ModelID
	// An older installation may already have an incomplete cached preview.
	legacy := filepath.Join(f.h.Config.Data, "models", id, "preview-plate-1.stl")
	if e := os.WriteFile(legacy, []byte("old disconnected preview"), 0600); e != nil {
		t.Fatal(e)
	}
	for n := 0; n < 2; n++ {
		w := f.call("GET", "/api/models/"+id+"/mesh?plate=1", nil, nil)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		mesh, e := model.ParseSTL(w.Body.Bytes())
		if e != nil || mesh.Min != (model.Vec{90, 90, 0}) || mesh.Max != (model.Vec{100, 100, 10}) {
			t.Fatal(mesh, e)
		}
	}
	w := f.call("GET", "/api/models/"+id+"/original", nil, nil)
	if w.Code != 200 || sha256.Sum256(w.Body.Bytes()) != sha256.Sum256(data) {
		t.Fatal("Preview changed original archive", w.Code)
	}
	for _, q := range []string{"", "?plate=0", "?plate=101", "?plate=2"} {
		w = f.call("GET", "/api/models/"+id+"/mesh"+q, nil, nil)
		if w.Code == 200 {
			t.Fatal("invalid plate accepted", q)
		}
	}
}

func TestDenseSTLPreviewKeepsEverySurfaceForExistingImports(t *testing.T) {
	f := setup(t)
	// Distinct translated closed cubes exceed the old 100k-facet cutoff.
	mesh := model.Mesh{}
	for n := 0; n < 8400; n++ {
		for _, tri := range model.Cube(1).Triangles {
			for i := range tri {
				tri[i][0] += float64(n % 100)
				tri[i][1] += float64(n / 100)
			}
			mesh.Triangles = append(mesh.Triangles, tri)
		}
	}
	data := mesh.Binary()
	p := f.upload(f.project(), data, "detailed-solid.stl")
	id := p.Objects[0].ModelID
	// Reproduce retained metadata and the legacy sampled file without relying
	// on a fresh upload to repair old saved projects.
	var saved Model
	var raw string
	f.h.DB.QueryRow("SELECT data FROM slicer_models WHERE id=?", id).Scan(&raw)
	if e := json.Unmarshal([]byte(raw), &saved); e != nil {
		t.Fatal(e)
	}
	saved.PreviewTriangles = 50400
	if _, e := f.h.DB.Exec("UPDATE slicer_models SET data=? WHERE id=?", encode(saved), id); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(f.h.Config.Data, "models", id, "preview.stl"), model.Cube(1).Binary(), 0600); e != nil {
		t.Fatal(e)
	}
	w := f.call("GET", "/api/models/"+id+"/mesh?surface=full-v2", nil, nil)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	preview, e := model.ParseSTL(w.Body.Bytes())
	if e != nil {
		t.Fatal(e)
	}
	if len(preview.Triangles) != len(mesh.Triangles) {
		t.Fatalf("surface facets dropped: got %d want %d", len(preview.Triangles), len(mesh.Triangles))
	}
	w = f.call("GET", "/api/models/"+id+"/original", nil, nil)
	if sha256.Sum256(w.Body.Bytes()) != sha256.Sum256(data) {
		t.Fatal("preview changed original source")
	}
	w = f.call("GET", "/api/models", nil, nil)
	var models []Model
	if e := json.Unmarshal(w.Body.Bytes(), &models); e != nil {
		t.Fatal(e)
	}
	if len(models) != 1 || models[0].PreviewTriangles != len(mesh.Triangles) {
		t.Fatal("retained metadata still describes incomplete preview", models)
	}
}
