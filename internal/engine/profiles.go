package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Values map[string]any
type Profile struct {
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	File   string `json:"-"`
	Values Values `json:"-"`
}
type Profiles struct {
	Entries map[string]Profile
	cache   map[string]Values
	mu      sync.Mutex
}
type Machine struct {
	Model      string       `json:"model"`
	Name       string       `json:"name"`
	Width      float64      `json:"width"`
	Depth      float64      `json:"depth"`
	Height     float64      `json:"height"`
	Polygon    [][2]float64 `json:"polygon"`
	Exclusions [][2]float64 `json:"exclusions"`
	Values     Values       `json:"-"`
}

func Text(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case bool:
		if x {
			return "1"
		}
		return "0"
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case []any:
		if len(x) > 0 {
			return Text(x[0])
		}
	}
	return ""
}
func Texts(v any) []string {
	var out []string
	switch x := v.(type) {
	case string:
		if x != "" {
			out = append(out, x)
		}
	case []any:
		for _, y := range x {
			out = append(out, Text(y))
		}
	}
	return out
}
func clone(v Values) Values {
	b, _ := json.Marshal(v)
	var out Values
	_ = json.Unmarshal(b, &out)
	return out
}

// Native resources are indexed by their declared names, including non-instantiated
// inheritance parents and G-code includes. Nothing resolves against uploaded files.
func LoadProfiles(root string) (*Profiles, error) {
	p := &Profiles{Entries: map[string]Profile{}, cache: map[string]Values{}}
	raw, e := os.ReadFile(filepath.Join(filepath.Dir(root), "BBL.json"))
	if e != nil {
		return nil, e
	}
	var manifest map[string]json.RawMessage
	if e = json.Unmarshal(raw, &manifest); e != nil {
		return nil, e
	}
	for _, kind := range []string{"machine", "process", "filament"} {
		var entries []struct {
			Name string `json:"name"`
			Path string `json:"sub_path"`
		}
		if e = json.Unmarshal(manifest[kind+"_list"], &entries); e != nil {
			return nil, e
		}
		if len(entries) == 0 {
			return nil, fmt.Errorf("Release manifest has no %s profiles", kind)
		}
		for _, item := range entries {
			rel := filepath.FromSlash(item.Path)
			if filepath.IsAbs(rel) || strings.Contains(item.Path, "..") || !strings.HasPrefix(item.Path, kind+"/") {
				return nil, errors.New("Unsafe release profile manifest path")
			}
			file := filepath.Join(root, rel)
			st, e := os.Lstat(file)
			if e != nil {
				return nil, e
			}
			if !st.Mode().IsRegular() || st.Size() > 4<<20 {
				return nil, errors.New("Invalid profile resource file")
			}
			data, e := os.ReadFile(file)
			if e != nil {
				return nil, e
			}
			var v Values
			if e = json.Unmarshal(data, &v); e != nil {
				return nil, e
			}
			if Text(v["name"]) != item.Name {
				return nil, fmt.Errorf("Profile manifest/name mismatch: %s", item.Name)
			}
			key := kind + "/" + item.Name
			if _, ok := p.Entries[key]; ok {
				return nil, fmt.Errorf("Duplicate release profile: %s", item.Name)
			}
			p.Entries[key] = Profile{Name: item.Name, Kind: kind, File: file, Values: v}
		}
	}
	return p, nil
}
func (p *Profiles) Resolve(kind, name string) (Values, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.resolve(kind, name, map[string]bool{})
}
func (p *Profiles) resolve(kind, name string, visiting map[string]bool) (Values, error) {
	key := kind + "/" + name
	if v, ok := p.cache[key]; ok {
		return clone(v), nil
	}
	if visiting[key] || len(visiting) > 40 {
		return nil, fmt.Errorf("Profile inheritance/include cycle: %s", name)
	}
	entry, ok := p.Entries[key]
	if !ok {
		return nil, fmt.Errorf("Missing release-matched %s profile/template: %s", kind, name)
	}
	visiting[key] = true
	defer delete(visiting, key)
	result := Values{}
	merge := func(parent string) error {
		v, e := p.resolve(kind, parent, visiting)
		if e != nil {
			return e
		}
		for k, x := range v {
			result[k] = x
		}
		return nil
	}
	for _, parent := range Texts(entry.Values["inherits"]) {
		if e := merge(parent); e != nil {
			return nil, e
		}
	}
	for _, include := range Texts(entry.Values["include"]) {
		if e := merge(include); e != nil {
			return nil, e
		}
	}
	for k, v := range entry.Values {
		if k != "inherits" && k != "include" {
			result[k] = v
		}
	}
	for _, key := range []string{"post_process", "post_process_script"} {
		if s := Texts(result[key]); len(s) > 0 && strings.Join(s, "") != "" {
			return nil, errors.New("Executable post-processing scripts are not allowed")
		}
	}
	p.cache[key] = clone(result)
	return result, nil
}
func (p *Profiles) Machine(model string) (Machine, error) {
	var m Machine
	m.Model = model
	switch model {
	case "A1":
		m.Name = "Bambu Lab A1 0.4 nozzle"
	case "A1 mini":
		m.Name = "Bambu Lab A1 mini 0.4 nozzle"
	default:
		return m, errors.New("Choose A1 or A1 mini with a confirmed 0.4 mm nozzle")
	}
	v, e := p.Resolve("machine", m.Name)
	if e != nil {
		return m, e
	}
	m.Values = v
	if Text(v["nozzle_diameter"]) != "0.4" {
		return m, errors.New("The machine profile does not match the confirmed nozzle")
	}
	for _, key := range []string{"machine_start_gcode", "machine_end_gcode", "change_filament_gcode", "layer_change_gcode", "time_lapse_gcode"} {
		if len(Text(v[key])) < 4 {
			return m, fmt.Errorf("Missing machine G-code template: %s", key)
		}
	}
	parse := func(key string) ([][2]float64, error) {
		var pts [][2]float64
		for _, s := range Texts(v[key]) {
			parts := strings.Split(s, "x")
			if len(parts) != 2 {
				return nil, fmt.Errorf("Invalid printable geometry: %s", key)
			}
			x, e := strconv.ParseFloat(parts[0], 64)
			if e != nil {
				return nil, e
			}
			y, e := strconv.ParseFloat(parts[1], 64)
			if e != nil {
				return nil, e
			}
			pts = append(pts, [2]float64{x, y})
		}
		return pts, nil
	}
	if m.Polygon, e = parse("printable_area"); e != nil {
		return m, e
	}
	if len(m.Polygon) < 3 {
		return m, errors.New("Machine has no precise printable polygon")
	}
	m.Exclusions, e = parse("bed_exclude_area")
	if e != nil {
		return m, e
	}
	minX, minY := m.Polygon[0][0], m.Polygon[0][1]
	var maxX, maxY float64
	for _, pt := range m.Polygon {
		minX = min(minX, pt[0])
		minY = min(minY, pt[1])
		maxX = max(maxX, pt[0])
		maxY = max(maxY, pt[1])
	}
	m.Width = maxX - minX
	m.Depth = maxY - minY
	m.Height, e = strconv.ParseFloat(Text(v["printable_height"]), 64)
	return m, e
}
func (p *Profiles) Compatible(kind, machine string) []Profile {
	var out []Profile
	for _, entry := range p.Entries {
		if entry.Kind != kind || Text(entry.Values["instantiation"]) != "true" {
			continue
		}
		v, e := p.Resolve(kind, entry.Name)
		if e != nil {
			continue
		}
		for _, name := range Texts(v["compatible_printers"]) {
			if name == machine {
				out = append(out, Profile{Name: entry.Name, Kind: kind})
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
func (p *Profiles) CompatibleValues(kind, name, machine string) (Values, error) {
	v, e := p.Resolve(kind, name)
	if e != nil {
		return nil, e
	}
	for _, x := range Texts(v["compatible_printers"]) {
		if x == machine {
			return v, nil
		}
	}
	return nil, fmt.Errorf("%s is not compatible with %s", name, machine)
}

func WriteJSON(path string, v any) error {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0600)
}
