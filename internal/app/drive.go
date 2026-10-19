package app

import (
	"encoding/json"
	"errors"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type DriveSave struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	ObjectID string `json:"objectID"`
	Name     string `json:"name"`
	State    string `json:"state"`
	SHA256   string `json:"sha256"`
	Bytes    int64  `json:"bytes"`
	Source   string `json:"source"`
	Path     string `json:"path"`
	Message  string `json:"message"`
	Lease    string `json:"lease,omitempty"`
	Until    int64  `json:"until,omitempty"`
}

func (a *App) queueDrive(user, kind, id, name string) {
	if a.H.Config.WorkspaceURL == "" {
		return
	}
	d := DriveSave{ID: apphost.ID(), Kind: kind, ObjectID: id, Name: cleanName(name), State: "pending", Message: "Waiting to save an ordinary copy in Drive while My Slicer is open"}
	if kind == "model" {
		m, e := a.getModel(user, id)
		if e != nil {
			return
		}
		d.SHA256 = m.SHA256
		d.Bytes = m.Size
	} else {
		s, e := a.getSlice(user, id)
		if e != nil || s.Result == nil {
			return
		}
		d.SHA256 = s.Result.SHA256
		d.Bytes = s.Result.Bytes
	}
	_, _ = a.H.DB.Exec("INSERT OR IGNORE INTO slicer_drive(id,user_id,kind,object_id,name,state,data) VALUES(?,?,?,?,?,?,?)", d.ID, user, d.Kind, d.ObjectID, d.Name, d.State, encode(d))
}
func (a *App) getDrive(user, id string) (DriveSave, error) {
	var d DriveSave
	var raw string
	if !validID(id) {
		return d, errMissing
	}
	if e := a.H.DB.QueryRow("SELECT data FROM slicer_drive WHERE id=? AND user_id=?", id, user).Scan(&raw); e != nil {
		return d, errMissing
	}
	e := json.Unmarshal([]byte(raw), &d)
	return d, e
}
func (a *App) driveSaves(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT data FROM slicer_drive WHERE user_id=? ORDER BY rowid DESC LIMIT 500", uid(r))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []DriveSave{}
	for rows.Next() {
		var raw string
		var d DriveSave
		if rows.Scan(&raw) == nil && json.Unmarshal([]byte(raw), &d) == nil {
			d.Lease = ""
			if d.State == "saving" && d.Until < time.Now().Unix() {
				d.State = "pending"
			}
			out = append(out, d)
		}
	}
	apphost.JSON(w, 200, out)
}
func (a *App) driveSave(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Action string `json:"action"`
		Lease  string `json:"lease"`
		Source string `json:"source"`
		Path   string `json:"path"`
	}
	if !apphost.Decode(w, r, &body) {
		return
	}
	a.storageMu.Lock()
	defer a.storageMu.Unlock()
	d, e := a.getDrive(uid(r), r.PathValue("id"))
	if e != nil {
		a.problem(w, e)
		return
	}
	if body.Action == "claim" {
		if d.State == "saved" {
			d.Lease = ""
			apphost.JSON(w, 200, d)
			return
		}
		if d.Lease != "" && d.Until > time.Now().Unix() {
			apphost.Problem(w, 409, "Another tab is saving this ordinary Drive copy")
			return
		}
		d.State = "saving"
		d.Lease = apphost.ID()
		d.Until = time.Now().Add(15 * time.Minute).Unix()
		d.Message = "Saving ordinary file to Drive"
	} else {
		if !apphost.Equal(body.Lease, d.Lease) || d.Until < time.Now().Unix() {
			apphost.Problem(w, 409, "Drive save lease expired; reconnect and retry")
			return
		}
		switch body.Action {
		case "target":
			if body.Source == "" || len(body.Source) > 180 || !strings.HasPrefix(body.Path, "/") || path.Clean(body.Path) != body.Path || strings.ContainsAny(body.Path, "\\\r\n\x00") || len(body.Path) > 1000 {
				a.problem(w, errors.New("Choose a valid ordinary Drive destination"))
				return
			}
			d.Source = body.Source
			d.Path = body.Path
		case "saved":
			if d.Source == "" || d.Path == "" {
				a.problem(w, errors.New("Choose a Drive destination before confirming storage"))
				return
			}
			d.State = "saved"
			d.Message = "Saved an ordinary copy to Drive: " + d.Source + d.Path
			d.Lease = ""
			d.Until = 0
		case "defer":
			d.State = "pending"
			d.Message = "Reconnect Drive to resume saving; source files remain unchanged"
			d.Lease = ""
			d.Until = 0
		case "failed":
			d.State = "error"
			d.Message = "Drive copy was not confirmed. Retry; existing files will not be replaced"
			d.Lease = ""
			d.Until = 0
		default:
			apphost.Problem(w, 400, "Choose a supported Drive save action")
			return
		}
	}
	if _, e = a.H.DB.Exec("UPDATE slicer_drive SET state=?,data=? WHERE id=? AND user_id=?", d.State, encode(d), d.ID, uid(r)); e != nil {
		a.H.DBError(w, e)
		return
	}
	apphost.JSON(w, 200, d)
}
func (a *App) driveFile(w http.ResponseWriter, r *http.Request) {
	d, e := a.getDrive(uid(r), r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, e.Error())
		return
	}
	var path string
	if d.Kind == "model" {
		m, e := a.getModel(uid(r), d.ObjectID)
		if e != nil {
			a.problem(w, e)
			return
		}
		path = filepath.Join(a.H.Config.Data, "models", m.ID, "original."+m.Kind)
	} else {
		s, e := a.getSlice(uid(r), d.ObjectID)
		if e != nil || s.State != "ready" {
			apphost.Problem(w, 404, "Completed native result unavailable")
			return
		}
		path = filepath.Join(a.H.Config.Data, "slices", s.ID, "result.gcode.3mf")
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="my-slicer-file"`)
	http.ServeFile(w, r, path)
}
