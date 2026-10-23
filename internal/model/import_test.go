package model

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativeProjectExtraTemperatureEntriesAreNotMissing(t *testing.T) {
	settings := map[string]any{"printer_model": "Bambu Lab P2S", "nozzle_diameter": []string{"0.4"}, "filament_settings_id": []string{"Matte", "Basic"}, "filament_type": []string{"PLA", "PLA"}, "nozzle_temperature": []string{"220", "220", "220", "220"}, "filament_diameter": []string{"1.75", "1.75"}, "filament_density": []string{"1.32", "1.26"}}
	raw, _ := json.Marshal(settings)
	p := previewFixture(t, map[string]string{"3D/3dmodel.model": previewPart, "Metadata/project_settings.config": string(raw)})
	before, _ := os.ReadFile(p)
	info, err := InspectProject(p)
	if err != nil || len(info.Filaments) != 2 || len(stringsOf(info.Settings["nozzle_temperature"])) != 4 {
		t.Fatal("complete expanded temperature vector rejected or truncated", info, err)
	}
	after, _ := os.ReadFile(p)
	if sha256.Sum256(before) != sha256.Sum256(after) {
		t.Fatal("original project modified")
	}
	for _, values := range [][]string{{"220"}, {"220", ""}} {
		settings["nozzle_temperature"] = values
		raw, _ = json.Marshal(settings)
		p = previewFixture(t, map[string]string{"3D/3dmodel.model": previewPart, "Metadata/project_settings.config": string(raw)})
		if _, err := InspectProject(p); err == nil {
			t.Fatal("genuinely missing/empty values accepted", values)
		}
	}
}

func TestArchiveReferenceAndScriptBoundaries(t *testing.T) {
	for _, data := range []string{`<model><component path="../../outside.model"/></model>`, `<Relationships><Relationship Target="file:///C:/private"/></Relationships>`, `<config><metadata key="post_process" value="run.cmd"/></config>`, `<model><component path="/3D/%2e%2e/model"/></model>`} {
		if e := validateReferences([]byte(data)); e == nil {
			t.Fatal("unsafe reference/script accepted", data)
		}
	}
	if e := validateReferences([]byte(`<model><component path="/3D/Objects/object_1.model"/></model>`)); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "unsafe.3mf")
	f, _ := os.Create(path)
	z := zip.NewWriter(f)
	w, _ := z.Create("3D/3dmodel.model")
	w.Write([]byte(`<model/>`))
	w, _ = z.Create("Metadata/project_settings.config")
	w.Write([]byte(`{"post_process":["run.cmd"]}`))
	z.Close()
	f.Close()
	if _, e := InspectProject(path); e == nil || !strings.Contains(e.Error(), "post-processing") {
		t.Fatal(e)
	}
}
func TestFullCircleNativeExtrusion(t *testing.T) {
	p, e := ParseGCode(strings.NewReader("; layer num/total_layer_count: 1/1\nG90\nM83\nG1 X10 Y10 Z0.2\nG2 X10 Y10 I5 J0 E2\n"))
	if e != nil || p.Extrusions < 10 {
		t.Fatal("full arc was discarded", p.Extrusions, e)
	}
}
