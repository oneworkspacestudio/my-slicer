package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSliceFailureUsesSelectedPlateWarning(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "result.json"), []byte(`{"return_code":-100,"error_string":"Failed slicing the model. Please verify all plates before uploading.","sliced_plates":[{"id":1,"warning_message":"overlap"},{"id":2,"warning_message":"Empty layer at height 1.2 in private-model.stl"}]}`), 0600)
	s := sliceFailure(dir, 2).Error()
	if !strings.Contains(s, "Plate 2:") || !strings.Contains(s, "smaller layer height") || strings.Contains(s, "overlap") || strings.Contains(s, "private-model") || strings.Contains(s, "upload") {
		t.Fatal(s)
	}
}

func TestSliceFailureRecovery(t *testing.T) {
	for _, tc := range []struct{ diagnostic, want string }{
		{"One of the plate is empty or has no object fully inside it", "do not fit"},
		{"G-code path conflict", "overlap"},
		{"std::bad_alloc", "not enough memory"},
		{"unknown private/path/model.stl", "Slicing could not finish"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			dir := t.TempDir()
			os.WriteFile(filepath.Join(dir, "engine.log"), []byte(tc.diagnostic), 0600)
			s := sliceFailure(dir, 1).Error()
			if !strings.Contains(s, tc.want) || strings.Contains(s, "private/path") {
				t.Fatal(s)
			}
		})
	}
}
