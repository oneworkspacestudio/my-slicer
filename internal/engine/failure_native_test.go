package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oneworkspacestudio/my-slicer/internal/model"
)

// This exercises the real CLI's -100 result and its explicit recovery, rather
// than treating a mocked failure message as native-engine acceptance.
func TestNativeEmptyLayersRecovery(t *testing.T) {
	root := os.Getenv("MYSLICER_TEST_NATIVE_ENGINE")
	if root == "" {
		t.Skip("explicit native engine and isolated NTFS fixture required")
	}
	base := filepath.ToSlash(os.Getenv("MYSLICER_TEST_NATIVE_ROOT"))
	if !strings.HasPrefix(strings.ToLower(base), "c:/projects/.cache/my-slicer-") {
		t.Fatal("choose a named isolated NTFS fixture")
	}
	dir, e := os.MkdirTemp(base, "empty-layers-")
	if e != nil {
		t.Fatal(e)
	}
	p, e := LoadProfiles(filepath.Join(root, "resources", "profiles", "BBL"))
	if e != nil {
		t.Fatal(e)
	}
	m, e := p.Machine("A1 mini")
	if e != nil {
		t.Fatal(e)
	}
	proc, e := p.Resolve("process", Text(m.Values["default_print_profile"]))
	if e != nil {
		t.Fatal(e)
	}
	fil, e := p.Resolve("filament", Text(m.Values["default_filament_profile"]))
	if e != nil {
		t.Fatal(e)
	}
	mesh := model.Mesh{}
	for _, z := range []float64{0, 5} {
		for _, tri := range model.Cube(10).Triangles {
			for i := range tri {
				tri[i][2] = tri[i][2]*0.1 + z
			}
			mesh.Triangles = append(mesh.Triangles, tri)
		}
	}
	mesh.Min, mesh.Max, e = model.Bounds(mesh.Triangles)
	if e != nil {
		t.Fatal(e)
	}
	input := filepath.Join(dir, "synthetic-gap.3mf")
	if e = model.WriteProject(input, []model.Object{{Mesh: mesh, Transform: model.Transform{X: 80, Y: 80, Scale: 1}}}); e != nil {
		t.Fatal(e)
	}
	c := SliceConfig{Machine: m.Values, Process: proc, Filaments: []Values{fil}, Plate: 1, Input: input, Overrides: Values{"enable_arc_fitting": "0", "print_sequence": "by layer", "curr_bed_type": "Textured PEI Plate", "enable_support": "0"}}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	_, e = Slice(ctx, root, filepath.Join(dir, "without-supports"), c)
	if e == nil || !strings.Contains(e.Error(), "smaller layer height") {
		t.Fatalf("expected actionable native empty-layer failure, got %v", e)
	}
	c.Overrides["enable_support"] = "1"
	r, e := Slice(ctx, root, filepath.Join(dir, "with-supports"), c)
	if e != nil || r.Layers < 2 || r.PreviewSegments < 1 {
		t.Fatalf("explicit support recovery failed: %+v %v", r, e)
	}
	t.Logf("Real empty-layer failure identified; explicit support recovery produced %d layers", r.Layers)
}
