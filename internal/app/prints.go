package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/oneworkspacestudio/my-slicer/internal/printer"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Print struct {
	ID              string               `json:"id"`
	PrinterID       string               `json:"printerID"`
	SliceID         string               `json:"sliceID"`
	State           string               `json:"state"`
	Message         string               `json:"message"`
	Done            int64                `json:"done"`
	Total           int64                `json:"total"`
	Created         string               `json:"created"`
	Updated         string               `json:"updated"`
	RequestedAt     string               `json:"requestedAt"`
	Spec            printer.PrintSpec    `json:"spec"`
	Options         printer.PrintOptions `json:"options"`
	Progress        *int                 `json:"progress"`
	Remaining       *int                 `json:"remaining"`
	CancelRequested bool                 `json:"cancelRequested"`
}

func taskID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return strconv.FormatUint(uint64(binary.LittleEndian.Uint32(b[:])&0x7fffffff|1), 10)
}
func (a *App) getPrint(user, id string) (Print, error) {
	var p Print
	if !validID(id) {
		return p, errMissing
	}
	var raw string
	e := a.H.DB.QueryRow("SELECT data FROM slicer_prints WHERE id=? AND user_id=?", id, user).Scan(&raw)
	if e != nil {
		return p, errMissing
	}
	e = json.Unmarshal([]byte(raw), &p)
	return p, e
}
func (a *App) printLock(id string) *sync.Mutex {
	m, _ := a.printLocks.LoadOrStore(id, &sync.Mutex{})
	return m.(*sync.Mutex)
}
func (a *App) storePrint(user string, p Print) error {
	lock := a.printLock(p.ID)
	lock.Lock()
	defer lock.Unlock()
	old, err := a.getPrint(user, p.ID)
	if err != nil {
		return err
	}
	terminal := func(s string) bool {
		return s == "completed" || s == "failed" || s == "cancelled" || s == "not_confirmed"
	}
	if terminal(old.State) {
		return nil
	}
	p.CancelRequested = p.CancelRequested || old.CancelRequested
	if p.Done < old.Done {
		p.Done = old.Done
	}
	if p.RequestedAt == "" {
		p.RequestedAt = old.RequestedAt
	}
	if (old.State == "printing" || old.State == "paused" || old.State == "preparing") && (p.State == "requesting_start" || p.State == "awaiting_printer" || p.State == "uploading") {
		p.State = old.State
		p.Message = old.Message
	}
	p.Updated = now()
	_, e := a.H.DB.Exec("UPDATE slicer_prints SET state=?,data=? WHERE id=? AND user_id=?", p.State, encode(p), p.ID, user)
	return e
}
func (a *App) prints(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT data FROM slicer_prints WHERE user_id=? ORDER BY created DESC LIMIT 100", uid(r))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	defer rows.Close()
	out := []Print{}
	for rows.Next() {
		var raw string
		var p Print
		if rows.Scan(&raw) == nil && json.Unmarshal([]byte(raw), &p) == nil {
			out = append(out, p)
		}
	}
	apphost.JSON(w, 200, out)
}
func (a *App) printRoute(w http.ResponseWriter, r *http.Request) {
	p, e := a.getPrint(uid(r), r.PathValue("id"))
	if e != nil {
		apphost.Problem(w, 404, e.Error())
		return
	}
	apphost.JSON(w, 200, p)
}
func (a *App) createPrint(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SliceID string               `json:"sliceID"`
		Confirm bool                 `json:"confirm"`
		Options printer.PrintOptions `json:"options"`
	}
	if !apphost.Decode(w, r, &body) {
		return
	}
	if !body.Confirm {
		apphost.Problem(w, 400, "Printing requires explicit confirmation of this sliced job and target printer")
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if !regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`).MatchString(key) {
		apphost.Problem(w, 400, "Use a unique dispatch confirmation key")
		return
	}
	requestHash := apphost.Hash(r.PathValue("id") + "\x00" + encode(body))
	a.mu.Lock()
	defer a.mu.Unlock()
	user := uid(r)
	var previousID, previousHash string
	if e := a.H.DB.QueryRow("SELECT id,request_hash FROM slicer_prints WHERE user_id=? AND idempotency=?", user, key).Scan(&previousID, &previousHash); e == nil {
		if previousHash != requestHash {
			apphost.Problem(w, 409, "This confirmation key already refers to a different job; review the new selection")
			return
		}
		previous, e := a.getPrint(user, previousID)
		if e != nil {
			a.problem(w, e)
			return
		}
		apphost.JSON(w, 200, previous)
		return
	}
	s, e := a.getSlice(user, body.SliceID)
	if e != nil || s.State != "ready" || s.Result == nil {
		apphost.Problem(w, 409, "Review a completed native slice before printing")
		return
	}
	project, e := a.getProject(user, s.ProjectID)
	if e != nil || project.Revision != s.Revision {
		apphost.Problem(w, 409, "Project geometry/settings changed after slicing. Slice the new revision first")
		return
	}
	record, _, e := a.getPrinter(user, r.PathValue("id"))
	if e != nil {
		a.problem(w, e)
		return
	}
	if s.Settings.PrinterID != record.ID || s.Settings.PrinterRevision != record.Revision || s.Settings.Model != record.Model || s.Settings.Nozzle != record.Nozzle || s.Settings.Plate != record.Plate {
		apphost.Problem(w, 409, "Printer configuration changed or differs from this slice. Select it and slice again")
		return
	}
	client, e := a.liveClient(user, record.ID)
	if e != nil {
		a.problem(w, e)
		return
	}
	id := apphost.ID()
	spec := printer.PrintSpec{Remote: "my-slicer-" + id + ".gcode.3mf", GCode: s.Result.GCode, MD5: s.Result.MD5, TaskID: taskID(), ModelID: s.Result.ModelID, FilamentSlots: s.Result.FilamentSlots}
	for _, m := range s.Result.Materials {
		spec.Materials = append(spec.Materials, printer.RequiredMaterial{ID: m.ID, Type: m.Type, Color: m.Color, NativeID: m.NativeID})
	}
	if _, e = printer.BuildStart(record, client.State(), spec, body.Options); e != nil {
		a.problem(w, e)
		return
	}
	p := Print{ID: id, PrinterID: record.ID, SliceID: s.ID, State: "validating", Message: "Checking the exact sliced archive and live target before upload", Total: s.Result.Bytes, Created: now(), Updated: now(), Spec: spec, Options: body.Options}
	if _, e = a.H.DB.Exec("INSERT INTO slicer_prints VALUES(?,?,?,?,?,?,?,?,?,?)", p.ID, user, p.PrinterID, p.SliceID, apphost.Hash(strings.ToUpper(record.Serial)), key, requestHash, p.State, encode(p), p.Created); e != nil {
		apphost.Problem(w, 409, "This physical printer already has an active or unresolved dispatch; reconcile it before another start")
		return
	}
	ctx := a.ctx
	a.wg.Add(1)
	go func() { defer a.wg.Done(); a.dispatch(ctx, user, p, s.Result.SHA256) }()
	a.H.Audit(user, "print-explicitly-confirmed", p.ID)
	apphost.JSON(w, 202, p)
}
func (a *App) dispatch(parent context.Context, user string, p Print, expectedSHA string) {
	a.H.Freeze.RLock()
	defer a.H.Freeze.RUnlock()
	ctx, cancel := context.WithTimeout(parent, 40*time.Minute)
	defer cancel()
	a.printCancels.Store(p.ID, cancel)
	defer a.printCancels.Delete(p.ID)
	fail := func(message string) { p.State = "failed"; p.Message = message; _ = a.storePrint(user, p) }
	client, e := a.liveClient(user, p.PrinterID)
	if e != nil {
		fail(e.Error())
		return
	}
	path := filepath.Join(a.H.Config.Data, "slices", p.SliceID, "result.gcode.3mf")
	f, e := os.Open(path)
	if e != nil {
		fail("The completed sliced archive is unavailable")
		return
	}
	hash := sha256.New()
	n, e := io.Copy(hash, f)
	f.Close()
	if e != nil || n != p.Total || hex.EncodeToString(hash.Sum(nil)) != expectedSHA {
		fail("The immutable sliced archive failed its checksum/size check; print start blocked")
		return
	}
	if _, e = printer.BuildStart(client.Record, client.State(), p.Spec, p.Options); e != nil {
		fail(e.Error())
		return
	}
	if e = client.TestFiles(ctx); e != nil {
		fail(e.Error())
		return
	}
	p.State = "uploading"
	p.Message = "Uploading the exact native sliced job to this printer; print has not started"
	_ = a.storePrint(user, p)
	e = client.Upload(ctx, path, p.Spec.Remote, func(done int64) {
		p.Done = done
		if done == p.Total {
			p.Message = "Bytes transferred; waiting for the printer to verify and finalize storage"
		}
		_ = a.storePrint(user, p)
	})
	if e != nil {
		fail(e.Error())
		return
	}
	// Commit uncertainty before the one and only start publish. A crash, timeout or
	// disconnect from this point can never cause an automatic second start.
	sent, e := a.beginStart(ctx, user, &p, client)
	if !sent {
		if e != nil {
			fail(e.Error())
		}
		return
	}
	if e != nil {
		p.State = "state_unknown"
		p.Message = "Start outcome is unknown. Reconnect and inspect the physical printer; no start is resent"
		_ = a.storePrint(user, p)
		return
	}

	p.State = "awaiting_printer"
	p.Message = "Start requested; waiting for matching printer-reported preparation/printing"
	_ = a.storePrint(user, p)
}
func (a *App) beginStart(ctx context.Context, user string, p *Print, client *printer.Client) (bool, error) {
	lock := a.printLock(p.ID)
	lock.Lock()
	defer lock.Unlock()
	old, e := a.getPrint(user, p.ID)
	if e != nil {
		return false, e
	}
	if old.State == "cancelled" || old.CancelRequested {
		return false, nil
	}
	if old.State != "uploading" {
		return false, errors.New("This dispatch is no longer eligible for a start; inspect its current state")
	}
	if e = ctx.Err(); e != nil {
		return false, e
	}
	p.State = "requesting_start"
	p.Message = "Upload verified. Requesting this exact print once"
	p.RequestedAt = now()
	p.Updated = now()
	if _, e = a.H.DB.Exec("UPDATE slicer_prints SET state=?,data=? WHERE id=? AND user_id=?", p.State, encode(p), p.ID, user); e != nil {
		return false, e
	}
	// Cancel-before-start and this publish share only this dispatch's lock. Other
	// printers and their status/controls remain independent.
	return true, client.Start(p.Spec, p.Options)
}
func (a *App) controlPrint(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Confirm bool `json:"confirm"`
	}
	if !apphost.Decode(w, r, &body) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	user := uid(r)
	p, e := a.getPrint(user, r.PathValue("id"))
	if e != nil {
		a.problem(w, e)
		return
	}
	action := filepath.Base(r.URL.Path)
	if action == "cancel" && body.Confirm {
		lock := a.printLock(p.ID)
		lock.Lock()
		current, err := a.getPrint(user, p.ID)
		if err == nil && (current.State == "validating" || current.State == "uploading") {
			current.State = "cancelled"
			current.CancelRequested = true
			current.Message = "Cancelled before any print start was requested"
			current.Updated = now()
			_, err = a.H.DB.Exec("UPDATE slicer_prints SET state=?,data=? WHERE id=? AND user_id=?", current.State, encode(current), current.ID, user)
			lock.Unlock()
			if err != nil {
				a.H.DBError(w, err)
				return
			}
			if c, ok := a.printCancels.Load(p.ID); ok {
				c.(context.CancelFunc)()
			}
			apphost.JSON(w, 202, current)
			return
		}
		lock.Unlock()
		p = current
	}
	client, e := a.liveClient(user, p.PrinterID)
	if e != nil {
		a.problem(w, e)
		return
	}
	state := client.State()
	if action == "reconcile" {
		p = a.reconcile(p, client)
		_ = a.storePrint(user, p)
		apphost.JSON(w, 200, p)
		return
	}
	if action == "resolve" {
		if !body.Confirm || p.State != "state_unknown" || !state.Idle() {
			apphost.Problem(w, 409, "Only an explicitly checked idle printer can resolve an unknown dispatch; verify the bed and live printer first")
			return
		}
		p.State = "not_confirmed"
		p.Message = "Outcome remained unconfirmed; user explicitly checked the idle printer and resolved this attempt. No print was resent"
		_ = a.storePrint(user, p)
		a.H.Audit(user, "ambiguous-dispatch-resolved-while-idle", p.ID)
		apphost.JSON(w, 200, p)
		return
	}
	if !body.Confirm {
		apphost.Problem(w, 400, "Confirm this action for the matching live printer job")
		return
	}
	if p.State == "completed" || p.State == "failed" || p.State == "cancelled" || p.State == "not_confirmed" {
		apphost.Problem(w, 409, "This dispatch has no active physical job to control")
		return
	}
	if e = client.Control(action, p.Spec, taskID()); e != nil {
		a.problem(w, e)
		return
	}
	if action == "cancel" {
		p.CancelRequested = true
	}
	p.Message = strings.Title(action) + " requested; waiting for actual printer state"
	_ = a.storePrint(user, p)
	a.H.Audit(user, "print-control-"+action, p.ID)
	apphost.JSON(w, 202, p)
}
func (a *App) reconcile(p Print, client *printer.Client) Print {
	s := client.State()
	if !s.Fresh() {
		if p.State != "uploading" && p.State != "validating" {
			p.State = "state_unknown"
			p.Message = "Live printer state unavailable. Reconnect to reconcile; no start/cancel command is sent automatically"
		}
		return p
	}
	if printer.Matches(s, p.Spec) {
		p.Progress = s.Progress
		p.Remaining = s.Remaining
		switch s.Status {
		case "PREPARE":
			p.State = "preparing"
			p.Message = "The matching printer job reports preparation"
		case "RUNNING":
			p.State = "printing"
			p.Message = "The printer reports this matching job running"
		case "PAUSE":
			p.State = "paused"
			p.Message = "The printer reports this matching job paused"
		case "FINISH":
			p.State = "completed"
			p.Message = "The printer reports this matching job finished"
		case "FAILED":
			if p.CancelRequested {
				p.State = "cancelled"
				p.Message = "The matching printer job stopped after your cancellation"
			} else {
				p.State = "failed"
				p.Message = "The matching printer job reports failure. Check its screen/error: " + s.PrintError
			}
		case "IDLE":
			if p.CancelRequested {
				p.State = "cancelled"
				p.Message = "The printer returned idle after your explicit cancellation"
			}
		}
		return p
	}
	if ack, ok := client.Ack(p.Spec.TaskID); ok && ack.Result != "" && ack.Result != "success" && ack.Result != "ok" {
		p.State = "failed"
		p.Message = "Printer rejected the start request: " + ack.Reason + ". Check its LAN/Developer mode, SD card and target configuration"
		return p
	}
	if t, e := time.Parse(time.RFC3339Nano, p.RequestedAt); e == nil && time.Since(t) > 35*time.Second && (p.State == "requesting_start" || p.State == "awaiting_printer") {
		p.State = "state_unknown"
		p.Message = "No matching physical job was confirmed. Reconcile live state; do not retry until the printer and bed are checked"
	}
	return p
}
func (a *App) monitor(ctx context.Context) {
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	lastRefresh := time.Time{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			a.H.Freeze.RLock()
			rows, e := a.H.DB.Query("SELECT id,user_id,data FROM slicer_printers")
			type entry struct{ ID, User, Data string }
			var devices []entry
			if e == nil {
				for rows.Next() {
					var d entry
					if rows.Scan(&d.ID, &d.User, &d.Data) == nil {
						devices = append(devices, d)
					}
				}
				rows.Close()
			}
			for _, d := range devices {
				var record printer.Record
				if json.Unmarshal([]byte(d.Data), &record) != nil || record.MQTTPin == "" {
					continue
				}
				device := a.device(d.ID)
				device.mu.Lock()
				client := device.Client
				busy := device.Busy
				device.mu.Unlock()
				if (client == nil || !client.State().Fresh()) && !busy {
					a.connectDevice(ctx, d.User, d.ID, false)
				} else if client != nil && time.Since(lastRefresh) > 15*time.Second {
					device.mu.Lock()
					if !device.Refreshing {
						device.Refreshing = true
						a.wg.Add(1)
						go func(d *Device, c *printer.Client) {
							defer a.wg.Done()
							defer func() { d.mu.Lock(); d.Refreshing = false; d.mu.Unlock() }()
							_ = c.Refresh()
						}(device, client)
					}
					device.mu.Unlock()
				}
			}
			if time.Since(lastRefresh) > 15*time.Second {
				lastRefresh = time.Now()
			}
			rows, e = a.H.DB.Query("SELECT user_id,data FROM slicer_prints WHERE state IN ('requesting_start','awaiting_printer','preparing','printing','paused','state_unknown')")
			type item struct {
				user string
				p    Print
			}
			var prints []item
			if e == nil {
				for rows.Next() {
					var user, raw string
					var p Print
					if rows.Scan(&user, &raw) == nil && json.Unmarshal([]byte(raw), &p) == nil {
						prints = append(prints, item{user, p})
					}
				}
				rows.Close()
			}
			for _, v := range prints {
				device := a.device(v.p.PrinterID)
				device.mu.Lock()
				client := device.Client
				device.mu.Unlock()
				if client != nil {
					p := a.reconcile(v.p, client)
					_ = a.storePrint(v.user, p)
				}
			}
			a.H.Freeze.RUnlock()
		}
	}
}
