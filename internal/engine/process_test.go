package engine

import "testing"

func TestProcessSchemaRejectsCommandTextAndNonFiniteValues(t *testing.T) {
	for _, values := range []map[string]string{
		{"outer_wall_speed": "60 --machine-start-gcode=M112"},
		{"machine_start_gcode": "G28"}, {"inner_wall_speed": "+Inf"},
	} {
		if ValidateProcess(values) == nil {
			t.Fatalf("unsafe native process arguments accepted: %v", values)
		}
	}
	if err := ValidateProcess(map[string]string{"outer_wall_speed": "60", "support_type": "tree(auto)", "initial_layer_print_height": "0.2"}); err != nil {
		t.Fatal(err)
	}
}
