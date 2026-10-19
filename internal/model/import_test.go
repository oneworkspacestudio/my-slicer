package model

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
