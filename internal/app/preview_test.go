package app

import (
	"crypto/sha256"
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
