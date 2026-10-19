package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/oneworkspacestudio/my-slicer/internal/printer"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"sync"
	"time"
)

func (a *App) getPrinter(user, id string) (printer.Record, []byte, error) {
	var record printer.Record
	var cipher []byte
	var raw string
	if !validID(id) {
		return record, nil, errMissing
	}
	e := a.H.DB.QueryRow("SELECT data,cipher FROM slicer_printers WHERE id=? AND user_id=?", id, user).Scan(&raw, &cipher)
	if e != nil {
		return record, nil, errMissing
	}
	e = json.Unmarshal([]byte(raw), &record)
	return record, cipher, e
}
func (a *App) device(id string) *Device {
	a.devicesMu.Lock()
	defer a.devicesMu.Unlock()
	d := a.devices[id]
	if d == nil {
		d = &Device{Diagnostics: []printer.Diagnostic{}}
		a.devices[id] = d
	}
	return d
}
func (a *App) background() context.Context { a.mu.Lock(); defer a.mu.Unlock(); return a.ctx }
func (a *App) printers(w http.ResponseWriter, r *http.Request) {
	rows, e := a.H.DB.Query("SELECT data FROM slicer_printers WHERE user_id=? LIMIT 16", uid(r))
	if e != nil {
		a.H.DBError(w, e)
		return
	}
	records := []printer.Record{}
	for rows.Next() {
		var raw string
		var p printer.Record
		if rows.Scan(&raw) == nil && json.Unmarshal([]byte(raw), &p) == nil {
			records = append(records, p)
		}
	}
	rows.Close()
	out := []any{}
	for _, record := range records {
		d := a.device(record.ID)
		d.mu.Lock()
		status := printer.State{Status: "UNKNOWN", Trays: []printer.Tray{}}
		if d.Client != nil {
			status = d.Client.State()
		}
		diag := append([]printer.Diagnostic{}, d.Diagnostics...)
		busy := d.Busy
		d.mu.Unlock()
		out = append(out, map[string]any{"config": record, "status": status, "diagnostics": diag, "testing": busy, "paired": record.MQTTPin != "" && record.FTPSPin != ""})
	}
	apphost.JSON(w, 200, out)
}
func (a *App) addPrinter(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config     printer.Record `json:"config"`
		AccessCode string         `json:"accessCode"`
	}
	if !apphost.Decode(w, r, &body) {
		return
	}
	body.Config.ID = apphost.ID()
	body.Config.Revision = 1
	body.Config.MQTTPin = ""
	body.Config.FTPSPin = ""
	if e := body.Config.Validate(); e != nil {
		a.problem(w, e)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if _, e := printer.Resolve(ctx, body.Config.Host); e != nil {
		a.problem(w, e)
		return
	}
	cipher, e := printer.Protect(body.AccessCode)
	body.AccessCode = ""
	if e != nil {
		a.problem(w, e)
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	var count int
	a.H.DB.QueryRow("SELECT COUNT(*) FROM slicer_printers WHERE user_id=?", uid(r)).Scan(&count)
	if count >= 16 {
		apphost.Problem(w, 413, "Limit of16 saved printers reached")
		return
	}
	if _, e = a.H.DB.Exec("INSERT INTO slicer_printers VALUES(?,?,?,?,?)", body.Config.ID, uid(r), 1, encode(body.Config), cipher); e != nil {
		a.H.DBError(w, e)
		return
	}
	a.H.Audit(uid(r), "printer-paired-pending", body.Config.ID)
	apphost.JSON(w, 201, body.Config)
}
func (a *App) patchPrinter(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Config     printer.Record `json:"config"`
		AccessCode string         `json:"accessCode"`
	}
	if !apphost.Decode(w, r, &body) {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	old, cipher, e := a.getPrinter(uid(r), r.PathValue("id"))
	if e != nil {
		a.problem(w, e)
		return
	}
	if old.Revision != body.Config.Revision {
		apphost.Problem(w, 409, "Reload this printer's current configuration before saving")
		return
	}
	var active int
	a.H.DB.QueryRow("SELECT COUNT(*) FROM slicer_prints WHERE printer_id=? AND state NOT IN ('completed','failed','cancelled','not_confirmed')", old.ID).Scan(&active)
	if active != 0 && (old.Serial != body.Config.Serial || old.Model != body.Config.Model || old.Nozzle != body.Config.Nozzle || old.Plate != body.Config.Plate) {
		apphost.Problem(w, 409, "Resolve the active print before changing its printer configuration")
		return
	}
	if e = body.Config.Validate(); e != nil {
		a.problem(w, e)
		return
	}
	if body.AccessCode != "" {
		cipher, e = printer.Protect(body.AccessCode)
		body.AccessCode = ""
		if e != nil {
			a.problem(w, e)
			return
		}
	}
	body.Config.ID = old.ID
	body.Config.Revision = old.Revision + 1
	body.Config.MQTTPin = old.MQTTPin
	body.Config.FTPSPin = old.FTPSPin
	if body.Config.Serial != old.Serial || body.Config.Model != old.Model {
		body.Config.MQTTPin = ""
		body.Config.FTPSPin = ""
		body.Config.ManualIdentity = false
	}
	if _, e = a.H.DB.Exec("UPDATE slicer_printers SET revision=?,data=?,cipher=? WHERE id=? AND user_id=?", body.Config.Revision, encode(body.Config), cipher, old.ID, uid(r)); e != nil {
		a.H.DBError(w, e)
		return
	}
	d := a.device(old.ID)
	d.mu.Lock()
	if d.Client != nil {
		d.Client.Close()
		d.Client = nil
	}
	d.Revision = body.Config.Revision
	d.mu.Unlock()
	apphost.JSON(w, 200, body.Config)
}
func (a *App) discover(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	devices, e := printer.Discover(ctx)
	if e != nil {
		apphost.Problem(w, 400, "Local discovery is unavailable; enter the printer's IP and serial manually")
		return
	}
	apphost.JSON(w, 200, devices)
}
func (a *App) probe(w http.ResponseWriter, r *http.Request) {
	record, _, e := a.getPrinter(uid(r), r.PathValue("id"))
	if e != nil {
		a.problem(w, e)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 22*time.Second)
	defer cancel()
	ip, e := printer.Resolve(ctx, record.Host)
	if e != nil {
		a.problem(w, e)
		return
	}
	var pair printer.Pairing
	var mqttErr, ftpsErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); pair.MQTT, mqttErr = printer.ProbeCertificate(ctx, ip, 8883) }()
	go func() { defer wg.Done(); pair.FTPS, ftpsErr = printer.ProbeCertificate(ctx, ip, 990) }()
	wg.Wait()
	if mqttErr != nil {
		a.problem(w, mqttErr)
		return
	}
	if ftpsErr != nil {
		a.problem(w, ftpsErr)
		return
	}
	if observed, e := printer.Observe(ctx, ip); e == nil {
		pair.Model = observed.Model
		pair.Serial = observed.Serial
		pair.Name = observed.Name
		if pair.Serial != record.Serial || pair.Model != record.Model {
			apphost.Problem(w, 409, "This IP advertises a different printer serial/model. Correct the saved device before trusting it")
			return
		}
	}
	token := apphost.ID()
	a.devicesMu.Lock()
	for key, v := range a.trust {
		if time.Now().After(v.Expires) {
			delete(a.trust, key)
		}
	}
	a.trust[token] = Trust{Printer: record.ID, Revision: record.Revision, Pairing: pair, Expires: time.Now().Add(3 * time.Minute)}
	a.devicesMu.Unlock()
	apphost.JSON(w, 200, map[string]any{"challenge": token, "certificates": pair, "message": "Confirm this IP and serial on your physical printer before trusting these device-specific certificates. LAN Only disconnects Bambu Handy/cloud functions; no firmware change is made."})
}
func (a *App) trustPrinter(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Challenge string `json:"challenge"`
		Confirm   bool   `json:"confirm"`
	}
	if !apphost.Decode(w, r, &body) {
		return
	}
	record, _, e := a.getPrinter(uid(r), r.PathValue("id"))
	if e != nil {
		a.problem(w, e)
		return
	}
	a.devicesMu.Lock()
	challenge, ok := a.trust[body.Challenge]
	if ok {
		delete(a.trust, body.Challenge)
	}
	a.devicesMu.Unlock()
	if !body.Confirm || !ok || challenge.Printer != record.ID || challenge.Revision != record.Revision || time.Now().After(challenge.Expires) {
		apphost.Problem(w, 409, "Pairing confirmation expired or the printer changed; read its certificates again")
		return
	}
	record.MQTTPin = challenge.Pairing.MQTT.SHA256
	record.FTPSPin = challenge.Pairing.FTPS.SHA256
	record.Revision++
	if _, e = a.H.DB.Exec("UPDATE slicer_printers SET revision=?,data=? WHERE id=? AND user_id=?", record.Revision, encode(record), record.ID, uid(r)); e != nil {
		a.H.DBError(w, e)
		return
	}
	a.H.Audit(uid(r), "printer-certificate-trusted", record.ID)
	apphost.JSON(w, 200, record)
}
func (a *App) testPrinter(w http.ResponseWriter, r *http.Request) {
	record, _, e := a.getPrinter(uid(r), r.PathValue("id"))
	if e != nil {
		a.problem(w, e)
		return
	}
	a.connectDevice(a.background(), uid(r), record.ID, true)
	apphost.JSON(w, 202, map[string]string{"state": "checking", "message": "Checking identity, MQTT authentication/status and FTPS listing; no print or filament-changing command is sent"})
}
func (a *App) connectDevice(ctx context.Context, user, id string, force bool) {
	d := a.device(id)
	d.mu.Lock()
	if d.Busy || !force && time.Since(d.LastAttempt) < 20*time.Second {
		d.mu.Unlock()
		return
	}
	d.Busy = true
	d.LastAttempt = time.Now()
	d.Diagnostics = []printer.Diagnostic{{Stage: "connection", State: "running", Message: "Checking private LAN address, identity and secure services"}}
	d.mu.Unlock()
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		defer func() { d.mu.Lock(); d.Busy = false; d.mu.Unlock() }()
		record, cipher, e := a.getPrinter(user, id)
		if e != nil {
			return
		}
		code, e := printer.Unprotect(cipher)
		if e != nil {
			d.mu.Lock()
			d.Diagnostics = []printer.Diagnostic{{Stage: "credentials", State: "error", Message: e.Error()}}
			d.mu.Unlock()
			return
		}
		client := printer.New(record, code)
		code = ""
		job, cancel := context.WithTimeout(ctx, 40*time.Second)
		defer cancel()
		e = client.Connect(job)
		diag := []printer.Diagnostic{}
		if e != nil {
			diag = append(diag, printer.Diagnostic{Stage: "identity/status", State: "error", Message: e.Error()})
		} else {
			diag = append(diag, printer.Diagnostic{Stage: "identity/status", State: "ok", Message: "Authenticated live status and saved printer identity matched"})
			if e = client.TestFiles(job); e != nil {
				diag = append(diag, printer.Diagnostic{Stage: "file transfer", State: "error", Message: e.Error()})
			} else {
				diag = append(diag, printer.Diagnostic{Stage: "file transfer", State: "ok", Message: "Implicit FTPS authentication and passive storage listing passed; upload is verified during an explicit Print"})
			}
			s := client.State()
			if !s.CommandMode(record) {
				diag = append(diag, printer.Diagnostic{Stage: "print commands", State: "attention", Message: "Check/confirm LAN Only and Developer Mode in the printer settings; support is not established by MQTT login alone"})
			} else {
				diag = append(diag, printer.Diagnostic{Stage: "print commands", State: "pending", Message: "Local command mode is enabled/confirmed. Actual print start is verified only after you explicitly choose Print"})
			}
		}
		d.mu.Lock()
		if d.Client != nil {
			d.Client.Close()
		}
		d.Client = client
		d.Revision = record.Revision
		d.Diagnostics = diag
		d.mu.Unlock()
	}()
}
func (a *App) liveClient(user, id string) (*printer.Client, error) {
	record, _, e := a.getPrinter(user, id)
	if e != nil {
		return nil, e
	}
	d := a.device(id)
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.Client == nil || d.Revision != record.Revision {
		return nil, errors.New("Connect and test this printer before printing")
	}
	s := d.Client.State()
	if !s.Fresh() {
		return nil, errors.New("Printer status is stale or disconnected; reconnect and reconcile before printing")
	}
	return d.Client, nil
}
