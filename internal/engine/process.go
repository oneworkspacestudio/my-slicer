package engine

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Only reviewed process settings may cross from the UI into the native engine.
// Machine limits, firmware commands and custom G-code are never accepted here.
var processRanges = map[string][2]float64{
	"initial_layer_print_height": {.08, .4},
	"line_width":                 {.2, 1.2}, "initial_layer_line_width": {.2, 1.2},
	"outer_wall_line_width": {.2, 1.2}, "inner_wall_line_width": {.2, 1.2},
	"top_surface_line_width": {.2, 1.2}, "sparse_infill_line_width": {.2, 1.2},
	"internal_solid_infill_line_width": {.2, 1.2}, "support_line_width": {.2, 1.2},
	"top_shell_layers": {0, 50}, "bottom_shell_layers": {0, 50},
	"outer_wall_speed": {1, 500}, "inner_wall_speed": {1, 500},
	"sparse_infill_speed": {1, 500}, "top_surface_speed": {1, 500},
	"initial_layer_speed": {1, 100}, "travel_speed": {1, 500},
	"support_threshold_angle": {0, 90}, "raft_layers": {0, 20},
	"support_top_z_distance": {0, 2}, "support_bottom_z_distance": {0, 2},
	"support_base_pattern_spacing": {0, 10}, "support_interface_top_layers": {0, 20},
	"support_object_xy_distance": {0, 10}, "brim_width": {0, 30},
}
var processEnums = map[string]string{
	"seam_position": "nearest|aligned|back|random",
	"support_type":  "normal(auto)|tree(auto)|normal(manual)|tree(manual)",
	"support_style": "default|grid|snug|tree_slim|tree_strong|tree_hybrid|tree_organic",
}

func ValidateProcess(values map[string]string) error {
	for key, value := range values {
		if options, ok := processEnums[key]; ok {
			valid := false
			for _, option := range strings.Split(options, "|") {
				if value == option {
					valid = true
				}
			}
			if !valid {
				return fmt.Errorf("Choose a supported %s", key)
			}
			continue
		}
		bounds, ok := processRanges[key]
		if !ok {
			return fmt.Errorf("Unsupported process setting: %s", key)
		}
		number, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) || number < bounds[0] || number > bounds[1] {
			return fmt.Errorf("%s must be between %g and %g", key, bounds[0], bounds[1])
		}
		if (strings.HasSuffix(key, "_layers")) && number != math.Trunc(number) {
			return fmt.Errorf("%s must be a whole number", key)
		}
	}
	return nil
}
