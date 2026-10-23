package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

// A user chooses whether to download/share this report. It deliberately excludes
// project/model files, owner/device identity, raw logs, credentials and G-code.
func (a *App) sliceTroubleshooting(w http.ResponseWriter, r *http.Request) {
	s, e := a.getSlice(uid(r), r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, e.Error())
		return
	}
	if s.State == "queued" || s.State == "processing" {
		apphost.Problem(w, 409, "Wait for slicing to finish before downloading its report")
		return
	}
	var native struct {
		Code   int    `json:"return_code"`
		Error  string `json:"error_string"`
		Plate  int    `json:"plate_index"`
		Plates []struct {
			ID      int    `json:"id"`
			Warning string `json:"warning_message"`
		} `json:"sliced_plates"`
	}
	dir := filepath.Join(a.H.Config.Data, "slices", s.ID)
	if f, e := os.Open(filepath.Join(dir, "result.json")); e == nil {
		b, _ := io.ReadAll(io.LimitReader(f, (1<<20)+1))
		f.Close()
		if len(b) <= 1<<20 {
			_ = json.Unmarshal(b, &native)
		}
	}
	// Redact the only filesystem prefix the worker can emit in its diagnostics.
	native.Error = strings.ReplaceAll(native.Error, a.H.Config.Data, "[app data]")
	for i := range native.Plates {
		native.Plates[i].Warning = strings.ReplaceAll(native.Plates[i].Warning, a.H.Config.Data, "[app data]")
	}
	settings := map[string]any{}
	b, _ := json.Marshal(s.Settings)
	_ = json.Unmarshal(b, &settings)
	delete(settings, "printerID")
	delete(settings, "printerRevision")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Disposition", `attachment; filename="my-slicer-troubleshooting.json"`)
	apphost.JSON(w, 200, map[string]any{"schema": 1, "appVersion": a.H.Config.Version, "engineVersion": "02.08.02.61", "state": s.State, "settings": settings, "native": native, "modelFilesIncluded": false, "signInDetailsIncluded": false})
}
