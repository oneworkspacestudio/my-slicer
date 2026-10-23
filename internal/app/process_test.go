package app

import (
	"encoding/json"
	"testing"
)

func TestProcessSettingsPersistWithProjectAndRejectUnsafeNativeOptions(t *testing.T) {
	f := setup(t)
	response := f.call("POST", "/api/projects", map[string]any{"name": "Process fixture", "draft": true}, nil)
	if response.Code != 201 {
		t.Fatal(response.Code, response.Body.String())
	}
	var project Project
	if err := json.Unmarshal(response.Body.Bytes(), &project); err != nil {
		t.Fatal(err)
	}
	project.Settings.Process = map[string]string{"outer_wall_speed": "60", "support_type": "tree(auto)", "line_width": "0.42"}
	response = f.call("PATCH", "/api/projects/"+project.ID, project, nil)
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	if err := json.Unmarshal(response.Body.Bytes(), &project); err != nil {
		t.Fatal(err)
	}
	read := f.call("GET", "/api/projects/"+project.ID, nil, nil)
	var restored Project
	if err := json.Unmarshal(read.Body.Bytes(), &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Settings.Process["outer_wall_speed"] != "60" || restored.Settings.Process["support_type"] != "tree(auto)" {
		t.Fatal("native process values were not retained")
	}
	for _, bad := range []map[string]string{
		{"machine_start_gcode": "G28"}, {"travel_speed": "NaN"}, {"outer_wall_speed": "501"},
		{"support_type": "made-up"}, {"raft_layers": "1.5"}, {"line_width": "0"},
	} {
		project.Settings.Process = bad
		response = f.call("PATCH", "/api/projects/"+project.ID, project, nil)
		if response.Code != 400 {
			t.Fatalf("unsafe or invalid settings accepted: %v: %d %s", bad, response.Code, response.Body.String())
		}
	}
	read = f.call("GET", "/api/projects/"+project.ID, nil, nil)
	if err := json.Unmarshal(read.Body.Bytes(), &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Revision != project.Revision || restored.Settings.Process["outer_wall_speed"] != "60" {
		t.Fatal("invalid save changed the previous usable project")
	}
}
