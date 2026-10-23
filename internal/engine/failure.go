package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The CLI's generic -100 message discards its actionable per-plate warning.
// Classify the private diagnostic without exposing file paths or model names.
func sliceFailure(dir string, selected int) error {
	var native struct {
		Code   int    `json:"return_code"`
		Error  string `json:"error_string"`
		Plates []struct {
			ID      int    `json:"id"`
			Warning string `json:"warning_message"`
		} `json:"sliced_plates"`
	}
	b, _ := os.ReadFile(filepath.Join(dir, "result.json"))
	if len(b) <= 1<<20 {
		_ = json.Unmarshal(b, &native)
	}
	var detail strings.Builder
	detail.WriteString(native.Error)
	for _, p := range native.Plates {
		if p.ID == selected {
			detail.WriteString("\n" + p.Warning)
		}
	}
	log, _ := os.ReadFile(filepath.Join(dir, "engine.log"))
	if len(log) <= 1<<20 {
		detail.Write(log)
	}
	s := strings.ToLower(detail.String())
	message := "Slicing could not finish. Your model is kept. Return to Prepare, check the selected printer, plate and print settings, then try again."
	switch {
	case strings.Contains(s, "empty layer"), strings.Contains(s, "empty gcode layer"), strings.Contains(s, "too thin"), strings.Contains(s, "faulty mesh"):
		message = "Some parts disappear between printed layers. In Prepare, try a smaller layer height or enable supports for raised parts. If it still fails, repair the model in its original app and reopen it."
	case strings.Contains(s, "overlap"), strings.Contains(s, "conflict"):
		message = "Parts or print paths overlap. Separate the objects or correct overlapping parts in the original project, then reopen it and slice again."
	case strings.Contains(s, "empty plate"), strings.Contains(s, "no object fully inside"), strings.Contains(s, "outside"), strings.Contains(s, "exceeds the height"):
		message = "This plate is empty or its models do not fit the selected printer. In Prepare, choose another imported plate or the printer the project was made for."
	case strings.Contains(s, "no layers"), strings.Contains(s, "no extrusions"), strings.Contains(s, "no printable"):
		message = "No printable layers were produced. In Prepare, check that the model rests on the plate and try a smaller layer height."
	case strings.Contains(s, "bad_alloc"), strings.Contains(s, "out of memory"), strings.Contains(s, "paging file"):
		message = "There was not enough memory to finish this plate. Close other large apps and try again, or split the model into smaller plates."
	}
	return fmt.Errorf("Plate %d: %s", selected, message)
}
