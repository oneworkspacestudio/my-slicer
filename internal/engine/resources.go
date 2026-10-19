package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The shipped CLI explicitly looks up *_full directories (BambuStudio.cpp).
// Generate these from this release's complete inheritance/include graph instead
// of supplying unresolved GUI leaf presets or generic preparation G-code.
func PrepareResources(root string, p *Profiles) error {
	base := filepath.Join(root, "resources", "profiles", "BBL")
	write := func(kind, name string, v any) error {
		if strings.Contains(name, "..") || strings.ContainsAny(name, ":\\") {
			return fmt.Errorf("Unsafe resource name")
		}
		target := filepath.Join(base, kind+"_full", filepath.FromSlash(name)+".json")
		if e := os.MkdirAll(filepath.Dir(target), 0700); e != nil {
			return e
		}
		return WriteJSON(target, v)
	}
	for _, kind := range []string{"machine", "process", "filament"} {
		if e := os.MkdirAll(filepath.Join(base, kind+"_full"), 0700); e != nil {
			return e
		}
	}
	for _, model := range []string{"A1", "A1 mini"} {
		m, e := p.Machine(model)
		if e != nil {
			return e
		}
		if e = write("machine", m.Name, m.Values); e != nil {
			return e
		}
		modelName := "Bambu Lab " + model
		raw, e := os.ReadFile(filepath.Join(base, "machine", modelName+".json"))
		if e != nil {
			return e
		}
		var def Values
		if e = json.Unmarshal(raw, &def); e != nil {
			return e
		}
		id := Text(def["model_id"])
		if id == "" {
			return fmt.Errorf("Machine model %s has no native model identity", model)
		}
		if e = write("machine", modelName, def); e != nil {
			return e
		}
		for _, kind := range []string{"process", "filament"} {
			for _, entry := range p.Compatible(kind, m.Name) {
				v, e := p.Resolve(kind, entry.Name)
				if e != nil {
					return e
				}
				if e = write(kind, entry.Name, v); e != nil {
					return e
				}
			}
		}
	}
	return nil
}
