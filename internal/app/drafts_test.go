package app

import (
	"encoding/json"
	"testing"

	"github.com/oneworkspacestudio/my-slicer/internal/model"
)

func TestDraftPlateImportThenOptionalProjectSave(t *testing.T) {
	f := setup(t)
	legacy := f.project()
	if legacy.Draft {
		t.Fatal("Existing named projects must remain saved projects")
	}
	w := f.call("POST", "/api/projects", map[string]any{"name": "cube", "draft": true}, nil)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var plate Project
	if e := json.Unmarshal(w.Body.Bytes(), &plate); e != nil || !plate.Draft {
		t.Fatal("Draft plate was not persisted", e)
	}
	plate = f.upload(plate, model.Cube(10).Binary(0), "cube.stl")
	if !plate.Draft || len(plate.Objects) != 1 {
		t.Fatal("Import must retain the recoverable unsaved plate", plate)
	}
	var restored Project
	w = f.call("GET", "/api/projects/"+plate.ID, nil, nil)
	if e := json.Unmarshal(w.Body.Bytes(), &restored); w.Code != 200 || e != nil || !restored.Draft || restored.Objects[0] != plate.Objects[0] {
		t.Fatal("Reload lost the draft geometry", w.Code, e)
	}
	plate.Draft = false
	plate.Name = "Reusable cube"
	w = f.call("PATCH", "/api/projects/"+plate.ID, plate, nil)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var saved Project
	json.Unmarshal(w.Body.Bytes(), &saved)
	if saved.ID != plate.ID || saved.Draft || saved.Name != plate.Name || saved.Revision != plate.Revision+1 || saved.Objects[0] != plate.Objects[0] {
		t.Fatal("Optional save must keep the same model and plate", saved)
	}
}
