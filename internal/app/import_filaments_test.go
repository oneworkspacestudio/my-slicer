package app

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"testing"
)

func TestMultiMaterialRetargetKeepsSlotsAndOriginal(t *testing.T) {
	f := setup(t)
	p := f.project()
	var source bytes.Buffer
	z := zip.NewWriter(&source)
	for name, data := range map[string]string{
		"3D/3dmodel.model":                 `<model/>`,
		"Metadata/project_settings.config": `{"printer_model":"Bambu Lab P2S","nozzle_diameter":["0.4"],"filament_settings_id":["Matte @P2S","Basic @P2S"],"filament_type":["PLA","PLA"],"nozzle_temperature":["220","220","220","220"],"filament_diameter":["1.75","1.75"],"filament_density":["1.32","1.26"]}`,
	} {
		w, e := z.Create(name)
		if e != nil {
			t.Fatal(e)
		}
		w.Write([]byte(data))
	}
	if e := z.Close(); e != nil {
		t.Fatal(e)
	}
	p = f.upload(p, source.Bytes(), "two-material.3mf")
	if p.Settings.PreserveImported || len(p.Settings.Filaments) != 2 || p.Settings.Filaments[0] != "" || p.Settings.Filaments[1] != "" {
		t.Fatal("retarget collapsed slots or silently chose materials", p.Settings)
	}
	m, e := f.a.getModel(f.session.User.ID, p.Objects[0].ModelID)
	if e != nil {
		t.Fatal(e)
	}
	if m.Import.Printer != "Bambu Lab P2S" || m.Import.Notice == "" {
		t.Fatal("missing source printer guidance")
	}
	original := f.call("GET", "/api/models/"+m.ID+"/original", nil, nil)
	if original.Code != 200 || !bytes.Equal(original.Body.Bytes(), source.Bytes()) {
		t.Fatal("original archive changed", original.Code)
	}
	p.Settings.Filaments = []string{"Bambu PLA Matte @BBL A1"}
	if w := f.call("PATCH", "/api/projects/"+p.ID, p, nil); w.Code != 400 {
		t.Fatal("single material fallback accepted", w.Code)
	}
	p.Settings.Filaments = []string{"Bambu PLA Matte @BBL A1", "Bambu PLA Basic @BBL A1"}
	w := f.call("PATCH", "/api/projects/"+p.ID, p, nil)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	w = f.call("GET", "/api/projects/"+p.ID, nil, nil)
	var restored Project
	if e = json.Unmarshal(w.Body.Bytes(), &restored); e != nil {
		t.Fatal(e)
	}
	if len(restored.Settings.Filaments) != 2 || restored.Settings.Filaments[1] != "Bambu PLA Basic @BBL A1" {
		t.Fatal("second slot lost on reload", restored.Settings)
	}
}
