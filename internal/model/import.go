package model

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

// The native reader extracts auxiliary members. Reject references outside the
// package before it sees an uploaded project, even when ZIP member paths are safe.
func validateReferences(data []byte) error {
	d := xml.NewDecoder(bytes.NewReader(data))
	for {
		token, e := d.Token()
		if e == io.EOF {
			return nil
		}
		if e != nil {
			return e
		}
		node, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		var key, value string
		for _, a := range node.Attr {
			if a.Name.Local == "key" {
				key = strings.ToLower(a.Value)
			}
			if a.Name.Local == "value" {
				value = a.Value
			}
		}
		if (strings.Contains(key, "post_process") || strings.Contains(key, "postprocessing")) && strings.TrimSpace(value) != "" {
			return errors.New("Executable post-processing metadata is not supported; remove it in the source project")
		}
		for _, a := range node.Attr {
			if a.Name.Local != "path" && a.Name.Local != "Target" {
				continue
			}
			v := a.Value
			if strings.ContainsAny(v, "\\:\x00") || strings.Contains(v, "%") || strings.Contains(v, "?") || strings.Contains(v, "#") {
				return errors.New("Unsupported external or encoded resource reference in 3MF")
			}
			for _, p := range strings.Split(v, "/") {
				if p == ".." || p == "." {
					return errors.New("3MF resource references must stay within the package")
				}
			}
		}
	}
}

type Plate struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}
type ImportInfo struct {
	Kind             string         `json:"kind"`
	Printer          string         `json:"printer"`
	Nozzles          []string       `json:"nozzles"`
	Filaments        []string       `json:"filaments"`
	MaterialTypes    []string       `json:"materialTypes"`
	Plates           []Plate        `json:"plates"`
	PreserveSettings bool           `json:"preserveSettings"`
	Notice           string         `json:"notice"`
	Settings         map[string]any `json:"-"`
}

func stringsOf(v any) []string {
	var out []string
	switch x := v.(type) {
	case string:
		out = append(out, x)
	case []any:
		for _, v := range x {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}
func emptyValue(v any) bool {
	if v == nil {
		return true
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x) == ""
	case []any:
		for _, v := range x {
			if !emptyValue(v) {
				return false
			}
		}
		return true
	}
	return false
}
func InspectProject(filename string) (ImportInfo, error) {
	out := ImportInfo{Kind: "3mf", Plates: []Plate{}, Notice: "Original project, plate membership, painting and modifiers are preserved. Browser transforms are available for STL objects; use Slice to inspect the exact selected plate toolpath."}
	z, files, e := OpenArchive(filename, 512<<20)
	if e != nil {
		return out, e
	}
	defer z.Close()
	if files["3D/3dmodel.model"] == nil {
		return out, errors.New("3MF has no supported standard3D/3dmodel.model entry")
	}
	for name, f := range files {
		ext := strings.ToLower(path.Ext(name))
		switch ext {
		case ".exe", ".dll", ".py", ".ps1", ".bat", ".cmd", ".sh":
			return out, errors.New("Executable payloads are not supported in model projects")
		}
		if ext == ".xml" || ext == ".rels" || ext == ".model" {
			b, e := ReadMember(f, 128<<20)
			if e != nil {
				return out, e
			}
			if e = SafeXML(b); e != nil {
				return out, e
			}
			if e = validateReferences(b); e != nil {
				return out, e
			}
			if ext == ".rels" {
				var rel struct {
					Relationships []struct {
						Target string `xml:"Target,attr"`
						Mode   string `xml:"TargetMode,attr"`
					} `xml:"Relationship"`
				}
				if e = xml.Unmarshal(b, &rel); e != nil {
					return out, e
				}
				for _, r := range rel.Relationships {
					if strings.EqualFold(r.Mode, "External") || strings.Contains(r.Target, "://") {
						return out, errors.New("Projects with external linked resources must be made self-contained before import")
					}
				}
			}
		}
	}
	if settings := files["Metadata/project_settings.config"]; settings != nil {
		b, e := ReadMember(settings, 16<<20)
		if e != nil {
			return out, e
		}
		if e = json.Unmarshal(b, &out.Settings); e != nil {
			return out, errors.New("Native project settings could not be read")
		}
		for key, v := range out.Settings {
			low := strings.ToLower(key)
			if (strings.Contains(low, "post_process") || strings.Contains(low, "postprocessing")) && !emptyValue(v) {
				return out, errors.New("Projects with executable post-processing scripts are blocked; remove the script in the source slicer and re-export")
			}
		}
		out.Printer, _ = out.Settings["printer_model"].(string)
		out.Nozzles = stringsOf(out.Settings["nozzle_diameter"])
		out.Filaments = stringsOf(out.Settings["filament_settings_id"])
		out.MaterialTypes = stringsOf(out.Settings["filament_type"])
		out.PreserveSettings = true
		if len(out.Nozzles) > 1 {
			return out, errors.New("This project requires multiple extruders/nozzles. Export a single-extruder A1/A1 mini project; the source file is unchanged")
		}
		if len(out.Filaments) > 16 {
			return out, errors.New("This project exceeds the supported16 filament slots")
		}
		if len(out.Filaments) == 0 || len(out.MaterialTypes) != len(out.Filaments) {
			return out, errors.New("Native project filament settings are incomplete; export a complete Bambu project")
		}
		for _, key := range []string{"nozzle_temperature", "filament_diameter", "filament_density"} {
			if len(stringsOf(out.Settings[key])) != len(out.Filaments) {
				return out, fmt.Errorf("Native project is missing complete%s values for every filament; no generic fallback is used", key)
			}
		}
	}
	if file := files["Metadata/model_settings.config"]; file != nil {
		b, e := ReadMember(file, 32<<20)
		if e != nil {
			return out, e
		}
		if e = SafeXML(b); e != nil {
			return out, e
		}
		var cfg struct {
			Plates []struct {
				Meta []struct {
					Key   string `xml:"key,attr"`
					Value string `xml:"value,attr"`
				} `xml:"metadata"`
			} `xml:"plate"`
		}
		if e = xml.Unmarshal(b, &cfg); e != nil {
			return out, e
		}
		if len(cfg.Plates) > 100 {
			return out, errors.New("Project plate count exceeds100")
		}
		for i, p := range cfg.Plates {
			display := fmt.Sprintf("Plate%d", i+1)
			index := i + 1
			for _, m := range p.Meta {
				if m.Key == "plate_name" && m.Value != "" {
					display = m.Value
				}
				if m.Key == "plater_id" {
					if n, e := strconv.Atoi(m.Value); e == nil && n >= 1 && n <= 100 {
						index = n
					}
				}
			}
			out.Plates = append(out.Plates, Plate{Index: index, Name: display})
		}
	}
	if len(out.Plates) == 0 {
		out.Plates = []Plate{{Index: 1, Name: "Plate1"}}
	}
	if !out.PreserveSettings {
		out.Notice = "Geometry-only3MF: select compatible A1/A1 mini printer, quality and filament profiles before slicing. The original archive is preserved."
	}
	return out, nil
}
