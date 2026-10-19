package printer

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type RequiredMaterial struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Color    string `json:"color"`
	NativeID string `json:"nativeID"`
}
type PrintSpec struct {
	Remote        string
	GCode         string
	MD5           string
	TaskID        string
	ModelID       string
	FilamentSlots int
	Materials     []RequiredMaterial
}
type PrintOptions struct {
	UseAMS                bool              `json:"useAMS"`
	Mapping               map[string]string `json:"mapping"`
	AcceptTrayDifferences bool              `json:"acceptTrayDifferences"`
	ExternalConfirmed     bool              `json:"externalConfirmed"`
	BedLeveling           bool              `json:"bedLeveling"`
}

func normalizedColor(s string) string {
	s = strings.ToUpper(strings.TrimPrefix(s, "#"))
	if len(s) > 6 {
		s = s[:6]
	}
	return s
}
func BuildStart(r Record, s State, spec PrintSpec, options PrintOptions) (map[string]any, error) {
	if !s.Idle() {
		return nil, errors.New("Printer is offline, busy or its live state is unknown; an existing print will not be overwritten")
	}
	if !s.IdentityVerified {
		return nil, errors.New("Verify this printer's physical identity before printing")
	}
	if !s.CommandMode(r) {
		return nil, errors.New("Local print commands are not enabled/confirmed. Check LAN Only and Developer Mode in this printer's settings")
	}
	if s.SDCard != nil && !*s.SDCard {
		return nil, errors.New("The printer reports no usable SD card; insert and prepare storage first")
	}
	model := "N2S"
	if r.Model == "A1 mini" {
		model = "N1"
	}
	if spec.ModelID != model {
		return nil, errors.New("This sliced plate targets a different printer model; switch target and slice again")
	}
	if !regexp.MustCompile(`^my-slicer-[a-f0-9]{48}\.gcode\.3mf$`).MatchString(spec.Remote) || !regexp.MustCompile(`^Metadata/plate_[1-9][0-9]*\.gcode$`).MatchString(spec.GCode) || !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(spec.MD5) {
		return nil, errors.New("The immutable sliced job has incomplete transfer metadata")
	}
	id, e := strconv.ParseUint(spec.TaskID, 10, 31)
	if e != nil || id == 0 {
		return nil, errors.New("Invalid unique dispatch identity")
	}
	if len(spec.Materials) == 0 || spec.FilamentSlots < 1 || spec.FilamentSlots > 16 {
		return nil, errors.New("Native plate filament metadata is missing or unsupported")
	}
	level := 0
	if options.BedLeveling {
		level = 1
	}
	command := map[string]any{"sequence_id": spec.TaskID, "command": "project_file", "param": spec.GCode, "url": "ftp://" + spec.Remote, "file": spec.Remote, "md5": spec.MD5, "bed_type": "auto", "timelapse": false, "bed_leveling": options.BedLeveling, "auto_bed_leveling": level, "flow_cali": false, "extrude_cali_flag": 0, "vibration_cali": false, "layer_inspect": false, "nozzle_offset_cali": 0, "use_ams": options.UseAMS, "subtask_name": strings.TrimSuffix(spec.Remote, ".gcode.3mf"), "project_id": spec.TaskID, "subtask_id": spec.TaskID, "task_id": spec.TaskID}
	if options.UseAMS {
		mapping := make([]int, spec.FilamentSlots)
		for i := range mapping {
			mapping[i] = -1
		} // Native unused-material marker, sized to the actual project slots.
		for _, material := range spec.Materials {
			if material.ID < 1 || material.ID > len(mapping) {
				return nil, errors.New("The plate filament ID is outside its real project settings")
			}
			key := options.Mapping[strconv.Itoa(material.ID)]
			var selected *Tray
			for i := range s.Trays {
				if s.Trays[i].Key == key {
					selected = &s.Trays[i]
					break
				}
			}
			if selected == nil || !selected.Available {
				return nil, fmt.Errorf("Choose a currently loaded AMS lite slot for native filament%d", material.ID)
			}
			if !strings.EqualFold(strings.TrimSpace(selected.Material), strings.TrimSpace(material.Type)) {
				return nil, fmt.Errorf("%s contains%s; this slice requires%s. Resolve the material profile before printing", selected.Label, selected.Material, material.Type)
			}
			if !options.AcceptTrayDifferences && (normalizedColor(selected.Color) != normalizedColor(material.Color) || selected.NativeID != "" && material.NativeID != "" && selected.NativeID != material.NativeID) {
				return nil, errors.New("Loaded tray color/preset differs from the sliced filament. Review the real loaded tray and confirm the difference, or slice with its matching profile")
			}
			mapping[material.ID-1] = selected.Wire
		}
		command["ams_mapping"] = mapping
	} else {
		if len(spec.Materials) != 1 {
			return nil, errors.New("A multi-material plate needs validated AMS lite mappings; slice a single-material plate to use an external spool")
		}
		if !options.ExternalConfirmed {
			return nil, errors.New("Confirm the required filament is physically loaded on this printer's external spool")
		}
	}
	return map[string]any{"print": command}, nil
}
func (c *Client) Start(spec PrintSpec, options PrintOptions) error {
	payload, e := BuildStart(c.Record, c.State(), spec, options)
	if e != nil {
		return e
	}
	return c.send(payload)
}
func Matches(s State, spec PrintSpec) bool {
	return s.TaskID == spec.TaskID && spec.TaskID != "" || s.Subtask == strings.TrimSuffix(spec.Remote, ".gcode.3mf") || path.Base(s.File) == spec.Remote
}
func (c *Client) Control(action string, spec PrintSpec, sequence string) error {
	s := c.State()
	if !s.Fresh() || !s.IdentityVerified || !Matches(s, spec) {
		return errors.New("Live printer state does not identify this job; reconnect/reconcile before controlling it")
	}
	command := ""
	switch action {
	case "pause":
		if s.Status == "RUNNING" {
			command = "pause"
		}
	case "resume":
		if s.Status == "PAUSE" {
			command = "resume"
		}
	case "cancel":
		if s.Status == "RUNNING" || s.Status == "PAUSE" || s.Status == "PREPARE" {
			command = "stop"
		}
	}
	if command == "" {
		return errors.New("This action is unavailable in the actual printer state")
	}
	return c.send(map[string]any{"print": map[string]any{"command": command, "sequence_id": sequence}})
}
