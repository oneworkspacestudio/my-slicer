package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProfileInheritanceIncludesPrecedenceAndCycles(t *testing.T) {
	p := &Profiles{Entries: map[string]Profile{}, cache: map[string]Values{}}
	for name, v := range map[string]Values{"parent": {"name": "parent", "machine_start_gcode": "official preparation", "temperature": "210"}, "template": {"name": "template", "machine_end_gcode": "official end"}, "child": {"name": "child", "inherits": "parent", "include": []any{"template"}, "temperature": "220"}} {
		p.Entries["machine/"+name] = Profile{Name: name, Kind: "machine", Values: v}
	}
	v, e := p.Resolve("machine", "child")
	if e != nil || v["temperature"] != "220" || v["machine_start_gcode"] != "official preparation" || v["machine_end_gcode"] != "official end" {
		t.Fatal(v, e)
	}
	p.Entries["machine/cycle"] = Profile{Values: Values{"inherits": "cycle"}}
	if _, e = p.Resolve("machine", "cycle"); e == nil {
		t.Fatal("Cycle accepted")
	}
	p.Entries["machine/missing"] = Profile{Values: Values{"include": []any{"missing-template"}}}
	if _, e = p.Resolve("machine", "missing"); e == nil {
		t.Fatal("Missing template accepted")
	}
}
func TestChecksumFailureDoesNotRunCode(t *testing.T) {
	p := filepath.Join(t.TempDir(), "not-an-engine.zip")
	os.WriteFile(p, []byte("wrong"), 0600)
	if e := checkHash(p, 5, Pin.SHA256); e == nil {
		t.Fatal("Bad hash accepted")
	}
}
func TestPinnedNativeInstallerAndBothRealProfiles(t *testing.T) {
	archive := os.Getenv("MYSLICER_TEST_ENGINE_ARCHIVE")
	if archive == "" {
		t.Skip("Set reviewed official archive to run native acceptance")
	}
	t.Setenv("MYSLICER_ENGINE_ARCHIVE", archive)
	m := NewManager(t.TempDir())
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	if !m.Install(ctx) {
		t.Fatal("Installer did not start")
	}
	for m.State().Busy {
		select {
		case <-ctx.Done():
			m.Wait()
			t.Fatal("Native install timed out")
		case <-time.After(200 * time.Millisecond):
		}
	}
	root, _, ready := m.Ready()
	if !ready {
		t.Fatal(m.State())
	}
	if _, e := os.Stat(filepath.Join(root, "BambuStudio.pdb")); !os.IsNotExist(e) {
		t.Fatal("Debug symbols were installed")
	}
	m.Check(ctx)
	for m.State().Busy {
		time.Sleep(100 * time.Millisecond)
	}
	if _, _, ready = m.Ready(); !ready {
		t.Fatal(m.State())
	}
}
