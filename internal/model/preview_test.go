package model

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func previewFixture(t *testing.T, members map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "assembly.3mf")
	f, e := os.Create(p)
	if e != nil {
		t.Fatal(e)
	}
	z := zip.NewWriter(f)
	for n, s := range members {
		w, e := z.Create(n)
		if e != nil {
			t.Fatal(e)
		}
		w.Write([]byte(s))
	}
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	f.Close()
	return p
}

const previewPart = `<model unit="millimeter"><resources><object id="1"><mesh><vertices><vertex x="0" y="0" z="0"/><vertex x="10" y="0" z="0"/><vertex x="0" y="10" z="5"/></vertices><triangles><triangle v1="0" v2="1" v3="2"/></triangles></mesh></object></resources></model>`

func TestPreviewProductionPartsTransformsAndPlateInstances(t *testing.T) {
	root := `<model unit="millimeter" xmlns:p="http://schemas.microsoft.com/3dmanufacturing/production/2015/06"><resources><object id="2"><components><component p:path="/3D/Objects/part.model" objectid="1" transform="0 1 0 -1 0 0 0 0 1 2 0 3"/></components></object></resources><build><item objectid="2" transform="1 0 0 0 1 0 0 0 1 90 90 0"/><item objectid="2" transform="1 0 0 0 1 0 0 0 1 300 90 0"/></build></model>`
	cfg := `<config><plate><metadata key="plater_id" value="1"/><model_instance><metadata key="object_id" value="2"/><metadata key="instance_id" value="0"/></model_instance></plate><plate><metadata key="plater_id" value="2"/><model_instance><metadata key="object_id" value="2"/><metadata key="instance_id" value="1"/></model_instance></plate></config>`
	p := previewFixture(t, map[string]string{"3D/3dmodel.model": root, "3D/Objects/part.model": previewPart, "Metadata/model_settings.config": cfg})
	for _, v := range []struct {
		plate  int
		lo, hi Vec
	}{{1, Vec{82, 90, 3}, Vec{92, 100, 8}}, {2, Vec{292, 90, 3}, Vec{302, 100, 8}}} {
		m, e := PreviewProject(p, v.plate)
		if e != nil || len(m.Triangles) != 1 || m.Min != v.lo || m.Max != v.hi {
			t.Fatal(v.plate, m, e)
		}
	}
	if _, e := PreviewProject(p, 3); e == nil {
		t.Fatal("missing plate accepted")
	}
}
func TestPreviewInvalidPartsAndUnits(t *testing.T) {
	for _, part := range []string{strings.Replace(previewPart, `v3="2"`, `v3="999"`, 1), `<model><resources><object id="1"><components><component objectid="1"/></components></object></resources><build><item objectid="1"/></build></model>`} {
		p := previewFixture(t, map[string]string{"3D/3dmodel.model": part})
		if _, e := PreviewProject(p, 1); e == nil {
			t.Fatal("invalid model preview accepted")
		}
	}
	p := previewFixture(t, map[string]string{"3D/3dmodel.model": strings.Replace(strings.Replace(previewPart, `unit="millimeter"`, `unit="inch"`, 1), `</model>`, `<build><item objectid="1"/></build></model>`, 1)})
	m, e := PreviewProject(p, 1)
	if e != nil || m.Max != (Vec{254, 254, 127}) {
		t.Fatal(m, e)
	}
}
