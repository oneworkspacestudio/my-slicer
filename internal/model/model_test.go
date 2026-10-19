package model

import (
	"archive/zip"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealMeshTransformsAndPrinterBounds(t *testing.T) {
	m := Cube(200).Centred()
	plate := [][2]float64{{0, 0}, {256, 0}, {256, 256}, {0, 256}}
	tr := Transform{X: 128, Y: 128, Scale: 1}
	if e := m.ValidatePlate(tr, plate, nil, 256); e != nil {
		t.Fatal(e)
	}
	mini := [][2]float64{{0, 0}, {180, 0}, {180, 180}, {0, 180}}
	if e := m.ValidatePlate(Transform{X: 90, Y: 90, Scale: 1}, mini, nil, 180); e == nil {
		t.Fatal("A1-only model accepted for mini")
	}
	tr.Scale = math.NaN()
	if tr.Valid() {
		t.Fatal("Nonfinite transform accepted")
	}
	roundtrip, e := ParseSTL(m.Binary(0))
	if e != nil || len(roundtrip.Triangles) != 12 {
		t.Fatal(e)
	}
	if _, e = ParseSTL([]byte("solid broken vertex NaN 1 2 endsolid")); e == nil {
		t.Fatal("Malformed STL accepted")
	}
}
func TestActualGCodeModesAndExtrusions(t *testing.T) {
	source := "G90\nM83\nG1 X10 Y10 Z.2\n; layer num/total_layer_count: 1/1\n; FEATURE: Outer wall\nG1 X20 E1\nG1 Y20 E1\nM82\nG92 E0\nG1 X10 E2\n"
	p, e := ParseGCode(strings.NewReader(source))
	if e != nil {
		t.Fatal(e)
	}
	if p.Extrusions != 3 || len(p.Layers) != 2 || p.Layers[1].Segments[2][3] != 10 {
		t.Fatal(p)
	}
	if _, e = ParseGCode(strings.NewReader("M83\nG1 X20 E1\n")); e == nil {
		t.Fatal("Preparation-only extrusion accepted as real printed layers")
	}
}
func TestArchivePathsAndXMLAreNotTrusted(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.3mf")
	f, _ := os.Create(p)
	z := zip.NewWriter(f)
	w, _ := z.Create("../escape.model")
	w.Write([]byte("not safe"))
	z.Close()
	f.Close()
	if a, _, e := OpenArchive(p, 1000); e == nil {
		a.Close()
		t.Fatal("Traversal accepted")
	}
	if e := SafeXML([]byte(`<!DOCTYPE model [<!ENTITY x SYSTEM "file:///secret">]><model>&x;</model>`)); e == nil {
		t.Fatal("Entity document accepted")
	}
}
