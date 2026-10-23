package engine

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/oneworkspacestudio/my-slicer/internal/model"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Material struct {
	ID       int     `json:"id"`
	Preset   string  `json:"preset"`
	Type     string  `json:"type"`
	Color    string  `json:"color"`
	NativeID string  `json:"nativeID"`
	Grams    float64 `json:"grams"`
}
type SliceResult struct {
	FilamentSlots   int        `json:"filamentSlots"`
	PrinterName     string     `json:"printerName"`
	PlateType       string     `json:"plateType"`
	Plate           int        `json:"plate"`
	GCode           string     `json:"gcode"`
	SHA256          string     `json:"sha256"`
	MD5             string     `json:"md5"`
	Bytes           int64      `json:"bytes"`
	ModelID         string     `json:"modelID"`
	Seconds         float64    `json:"seconds"`
	Grams           float64    `json:"grams"`
	Layers          int        `json:"layers"`
	Features        []string   `json:"features"`
	Materials       []Material `json:"materials"`
	PreviewSegments int        `json:"previewSegments"`
	Engine          string     `json:"engine"`
}
type SliceConfig struct {
	PreserveImported bool
	Overrides        Values
	Machine          Values
	Process          Values
	Filaments        []Values
	Plate            int
	Input            string
}
type LimitedLog struct {
	File  *os.File
	Bytes int64
}

func (l *LimitedLog) Write(b []byte) (int, error) {
	n := len(b)
	if l.Bytes < 1<<20 {
		part := b
		if int64(len(b))+l.Bytes > 1<<20 {
			part = b[:(1<<20)-l.Bytes]
		}
		_, e := l.File.Write(part)
		if e != nil {
			return 0, e
		}
		l.Bytes += int64(len(part))
	}
	return n, nil
}
func Slice(ctx context.Context, root, dir string, c SliceConfig) (SliceResult, error) {
	var result SliceResult
	if c.Plate < 1 || c.Plate > 100 || (!c.PreserveImported && len(c.Filaments) == 0) || len(c.Filaments) > 16 {
		return result, errors.New("Choose one plate and resolve every filament before slicing")
	}
	if e := os.MkdirAll(dir, 0700); e != nil {
		return result, e
	}
	configs := map[string]any{"machine.json": c.Machine}
	if c.Process != nil {
		configs["process.json"] = c.Process
	}
	for filename, v := range configs {
		if e := WriteJSON(filepath.Join(dir, filename), v); e != nil {
			return result, e
		}
	}
	var filaments []string
	for i, v := range c.Filaments {
		p := filepath.Join(dir, fmt.Sprintf("filament-%d.json", i+1))
		if e := WriteJSON(p, v); e != nil {
			return result, e
		}
		filaments = append(filaments, p)
	}
	settings := filepath.Join(dir, "machine.json")
	if c.Process != nil {
		settings += ";" + filepath.Join(dir, "process.json")
	}
	args := []string{"--debug", "2", "--arrange", "0", "--load-settings", settings}
	if len(filaments) > 0 {
		args = append(args, "--load-filaments", strings.Join(filaments, ";"))
	}
	allowed := map[string]bool{"layer_height": true, "sparse_infill_density": true, "wall_loops": true, "enable_support": true, "brim_type": true, "curr_bed_type": true, "enable_arc_fitting": true, "print_sequence": true}
	process := map[string]string{}
	for key, value := range c.Overrides {
		if !allowed[key] {
			process[key] = Text(value)
		}
	}
	if e := ValidateProcess(process); e != nil {
		return result, e
	}
	for key, value := range c.Overrides {
		args = append(args, "--"+strings.ReplaceAll(key, "_", "-")+"="+Text(value))
	}
	args = append(args, "--slice", strconv.Itoa(c.Plate), "--outputdir", dir, "--export-3mf", "result.gcode.3mf", "--export-settings", filepath.Join(dir, "effective.json"), c.Input)

	log, e := os.OpenFile(filepath.Join(dir, "engine.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if e != nil {
		return result, e
	}
	runErr := Run(ctx, filepath.Join(root, "bambu-studio.exe"), dir, args, &LimitedLog{File: log})
	log.Close()
	if runErr != nil {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		var native struct {
			Error string `json:"error_string"`
		}
		if b, e := os.ReadFile(filepath.Join(dir, "result.json")); e == nil {
			_ = json.Unmarshal(b, &native)
		}
		if native.Error != "" {
			return result, fmt.Errorf("Native slicing failed: %s", strings.ReplaceAll(native.Error, dir, "[slice folder]"))
		}
		return result, fmt.Errorf("Native slicing stopped (%v). Check the engine installation and Windows runtime libraries", runErr)
	}
	return InspectResult(dir, c.Plate)
}
func InspectResult(dir string, selected int) (SliceResult, error) {
	var result SliceResult
	filename := filepath.Join(dir, "result.gcode.3mf")
	z, files, e := model.OpenArchive(filename, 512<<20)
	if e != nil {
		return result, e
	}
	defer z.Close()
	path := fmt.Sprintf("Metadata/plate_%d.gcode", selected)
	f := files[path]
	if f == nil || f.UncompressedSize64 < 1000 {
		return result, errors.New("The selected plate has no complete native G-code output")
	}
	info, e := model.ReadMember(files["Metadata/slice_info.config"], 8<<20)
	if e != nil {
		return result, e
	}
	if e = model.SafeXML(info); e != nil {
		return result, e
	}
	var metadata struct {
		Plates []struct {
			Meta []struct {
				Key   string `xml:"key,attr"`
				Value string `xml:"value,attr"`
			} `xml:"metadata"`
			Filaments []struct {
				ID       int     `xml:"id,attr"`
				Type     string  `xml:"type,attr"`
				Color    string  `xml:"color,attr"`
				NativeID string  `xml:"tray_info_idx,attr"`
				Grams    float64 `xml:"used_g,attr"`
			} `xml:"filament"`
		} `xml:"plate"`
	}
	if e = xml.Unmarshal(info, &metadata); e != nil {
		return result, e
	}
	found := false
	for _, p := range metadata.Plates {
		values := map[string]string{}
		for _, v := range p.Meta {
			values[v.Key] = v.Value
		}
		if values["index"] != strconv.Itoa(selected) {
			continue
		}
		if values["outside"] == "true" {
			return result, errors.New("The native engine reports objects outside the printable area")
		}
		result.ModelID = values["printer_model_id"]
		result.Seconds, _ = strconv.ParseFloat(values["prediction"], 64)
		result.Grams, _ = strconv.ParseFloat(values["weight"], 64)
		for _, m := range p.Filaments {
			result.Materials = append(result.Materials, Material{ID: m.ID, Type: m.Type, Color: m.Color, NativeID: m.NativeID, Grams: m.Grams})
		}
		found = true
		break
	}
	if !found || result.ModelID == "" || len(result.Materials) == 0 || result.Seconds <= 0 {
		return result, errors.New("Native plate metadata is incomplete; the job cannot be printed")
	}
	configuration, e := model.ReadMember(files["Metadata/project_settings.config"], 16<<20)
	if e != nil {
		return result, e
	}
	var full Values
	if e = json.Unmarshal(configuration, &full); e != nil {
		return result, e
	}
	result.FilamentSlots = len(Texts(full["filament_settings_id"]))
	result.PrinterName = Text(full["printer_settings_id"])
	result.PlateType = Text(full["curr_bed_type"])
	if result.FilamentSlots < 1 || result.FilamentSlots > 16 {
		return result, errors.New("Native project filament slots are missing or exceed the supported limit")
	}
	reader, e := f.Open()
	if e != nil {
		return result, e
	}
	toolpath, e := model.ParseGCode(reader)
	reader.Close()
	if e != nil {
		return result, e
	}
	if e = os.MkdirAll(filepath.Join(dir, "preview"), 0700); e != nil {
		return result, e
	}
	for _, layer := range toolpath.Layers {
		if e = WriteJSON(filepath.Join(dir, "preview", fmt.Sprintf("layer-%d.json", layer.Number)), layer); e != nil {
			return result, e
		}
	}
	result.Plate = selected
	result.GCode = path
	result.Layers = len(toolpath.Layers) - 1
	result.Features = toolpath.Features
	result.PreviewSegments = toolpath.Extrusions
	result.Engine = "02.08.02.61"
	file, e := os.Open(filename)
	if e != nil {
		return result, e
	}
	defer file.Close()
	s, h := sha256.New(), md5.New()
	result.Bytes, e = io.Copy(io.MultiWriter(s, h), file)
	if e != nil {
		return result, e
	}
	result.SHA256 = hex.EncodeToString(s.Sum(nil))
	result.MD5 = hex.EncodeToString(h.Sum(nil))
	if e = WriteJSON(filepath.Join(dir, "summary.json"), result); e != nil {
		return result, e
	}
	return result, nil
}
func Validate(ctx context.Context, root, dir string, notify func(string)) error {
	p, e := LoadProfiles(filepath.Join(root, "resources", "profiles", "BBL"))
	if e != nil {
		return e
	}
	if e = PrepareResources(root, p); e != nil {
		return e
	}
	var receipts []SliceResult
	for _, name := range []string{"A1", "A1 mini"} {
		notify("Testing real native slice for " + name)
		m, e := p.Machine(name)
		if e != nil {
			return e
		}
		proc, e := p.Resolve("process", Text(m.Values["default_print_profile"]))
		if e != nil {
			return e
		}
		fil, e := p.Resolve("filament", Text(m.Values["default_filament_profile"]))
		if e != nil {
			return e
		}
		proc["enable_arc_fitting"] = "0"
		proc["print_sequence"] = "by layer"
		folder := filepath.Join(dir, name)
		if e = os.MkdirAll(folder, 0700); e != nil {
			return e
		}
		input := filepath.Join(folder, "input.3mf")
		if e = model.WriteProject(input, []model.Object{{Mesh: model.Cube(10).Centred(), Transform: model.Transform{X: m.Width / 2, Y: m.Depth / 2, Scale: 1}}}); e != nil {
			return e
		}
		jobCtx, cancel := context.WithTimeout(ctx, 4*time.Minute)
		receipt, e := Slice(jobCtx, root, folder, SliceConfig{Machine: m.Values, Process: proc, Filaments: []Values{fil}, Plate: 1, Input: input})
		cancel()
		if e != nil {
			return fmt.Errorf("%s: %w", name, e)
		}
		want := "N2S"
		if name == "A1 mini" {
			want = "N1"
		}
		if receipt.ModelID != want {
			return errors.New("Native output targets a different printer model")
		}
		receipts = append(receipts, receipt)
	}
	return WriteJSON(filepath.Join(dir, "validation.json"), receipts)
}
