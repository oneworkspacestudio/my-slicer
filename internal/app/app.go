package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/oneworkspacestudio/my-slicer/internal/engine"
	"github.com/oneworkspacestudio/my-slicer/internal/printer"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

type App struct {
	H           *apphost.Host
	Engine      *engine.Manager
	mu          sync.Mutex
	storageMu   sync.Mutex
	printMu     sync.Mutex
	ctx         context.Context
	cancelSlice map[string]context.CancelFunc
	devicesMu   sync.Mutex
	devices     map[string]*Device
	trust       map[string]Trust
	wg          sync.WaitGroup
}
type Trust struct {
	Printer  string
	Revision int
	Pairing  printer.Pairing
	Expires  time.Time
}
type Device struct {
	mu          sync.Mutex
	Client      *printer.Client
	Revision    int
	Busy        bool
	Refreshing  bool
	LastAttempt time.Time
	Diagnostics []printer.Diagnostic
}
type Settings struct {
	PrinterID        string          `json:"printerID"`
	PrinterRevision  int             `json:"printerRevision"`
	Model            string          `json:"model"`
	Nozzle           string          `json:"nozzle"`
	Plate            string          `json:"plate"`
	Quality          string          `json:"quality"`
	Filaments        []string        `json:"filaments"`
	NativePlate      int             `json:"nativePlate"`
	PreserveImported bool            `json:"preserveImported"`
	LayerHeight      float64         `json:"layerHeight"`
	Infill           int             `json:"infill"`
	Walls            int             `json:"walls"`
	Supports         bool            `json:"supports"`
	Brim             string          `json:"brim"`
	Overrides        map[string]bool `json:"overrides"`
}
type Instance struct {
	ID        string `json:"id"`
	ModelID   string `json:"modelID"`
	Transform any    `json:"-"`
}
type Project struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Revision int      `json:"revision"`
	Settings Settings `json:"settings"`
	Objects  []Object `json:"objects"`
	Updated  string   `json:"updated"`
}
type Slice struct {
	ID        string              `json:"id"`
	ProjectID string              `json:"projectID"`
	Revision  int                 `json:"revision"`
	State     string              `json:"state"`
	Message   string              `json:"message"`
	Created   string              `json:"created"`
	Updated   string              `json:"updated"`
	Result    *engine.SliceResult `json:"result"`
	Settings  Settings            `json:"settings"`
}

func New(h *apphost.Host) (*App, error) {
	if e := h.Migrate(1, `CREATE TABLE slicer_models(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,data TEXT NOT NULL,bytes INTEGER NOT NULL,created TEXT NOT NULL);
CREATE TABLE slicer_projects(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,revision INTEGER NOT NULL,data TEXT NOT NULL,updated TEXT NOT NULL);
CREATE TABLE slicer_slices(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,project_id TEXT NOT NULL,revision INTEGER NOT NULL,state TEXT NOT NULL,data TEXT NOT NULL,created TEXT NOT NULL);
CREATE INDEX slicer_slice_queue ON slicer_slices(state,created);
CREATE TABLE slicer_printers(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,revision INTEGER NOT NULL,data TEXT NOT NULL,cipher BLOB NOT NULL);
CREATE TABLE slicer_prints(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,printer_id TEXT NOT NULL,slice_id TEXT NOT NULL,physical TEXT NOT NULL,idempotency TEXT NOT NULL,request_hash TEXT NOT NULL,state TEXT NOT NULL,data TEXT NOT NULL,created TEXT NOT NULL,UNIQUE(user_id,idempotency));
CREATE UNIQUE INDEX slicer_active_printer ON slicer_prints(physical) WHERE state NOT IN ('completed','failed','cancelled','not_confirmed');
CREATE TABLE slicer_drive(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,kind TEXT NOT NULL,object_id TEXT NOT NULL,name TEXT NOT NULL,state TEXT NOT NULL DEFAULT 'pending',data TEXT NOT NULL DEFAULT '{}',UNIQUE(user_id,kind,object_id));`); e != nil {
		return nil, e
	}
	for _, dir := range []string{"models", "projects", "slices", "logs", "config"} {
		if e := os.MkdirAll(filepath.Join(h.Config.Data, dir), 0700); e != nil {
			return nil, e
		}
	}
	a := &App{H: h, Engine: engine.NewManager(h.Config.Data), ctx: context.Background(), cancelSlice: map[string]context.CancelFunc{}, devices: map[string]*Device{}, trust: map[string]Trust{}}
	a.Engine.FileLock = &h.Freeze
	h.DB.Exec(`UPDATE slicer_slices SET state='cancelled',data=json_set(data,'$.state','cancelled','$.message','Server restarted during preparation; original files are preserved. Slice again when ready.') WHERE state IN ('queued','processing')`)
	h.DB.Exec(`UPDATE slicer_prints SET state='failed',data=json_set(data,'$.state','failed','$.message','Interrupted before requesting print start; no automatic retry was made.') WHERE state IN ('validating','uploading')`)
	h.DB.Exec(`UPDATE slicer_prints SET state='state_unknown',data=json_set(data,'$.state','state_unknown','$.message','Reconnect to the printer to reconcile this job. No print start or cancellation is sent automatically.') WHERE state NOT IN ('completed','failed','cancelled','not_confirmed','state_unknown')`)
	m := h.Mux
	m.HandleFunc("GET /api/system", h.Require(a.system))
	m.HandleFunc("POST /api/setup/engine", h.Owner(a.installEngine))
	m.HandleFunc("POST /api/setup/engine/cancel", h.Owner(func(w http.ResponseWriter, r *http.Request) {
		a.Engine.Cancel()
		apphost.JSON(w, 202, a.Engine.State())
	}))
	m.HandleFunc("GET /api/profiles", h.Require(a.profiles))
	m.HandleFunc("GET /api/projects", h.Require(a.projects))
	m.HandleFunc("POST /api/projects", h.Require(a.createProject))
	m.HandleFunc("GET /api/projects/{id}", h.Require(a.projectRoute))
	m.HandleFunc("PATCH /api/projects/{id}", h.Require(a.updateProject))
	m.HandleFunc("POST /api/projects/{id}/models", h.Require(a.uploadModel))
	m.HandleFunc("POST /api/projects/{id}/arrange", h.Require(a.arrange))
	m.HandleFunc("GET /api/models", h.Require(a.models))
	m.HandleFunc("GET /api/models/{id}/original", h.Require(a.modelFile))
	m.HandleFunc("GET /api/models/{id}/mesh", h.Require(a.modelFile))
	m.HandleFunc("GET /api/slices", h.Require(a.slices))
	m.HandleFunc("POST /api/projects/{id}/slices", h.Require(a.createSlice))
	m.HandleFunc("GET /api/slices/{id}", h.Require(a.sliceRoute))
	m.HandleFunc("DELETE /api/slices/{id}", h.Require(a.deleteSlice))
	m.HandleFunc("POST /api/slices/{id}/cancel", h.Require(a.cancelSliceRoute))
	m.HandleFunc("GET /api/slices/{id}/preview", h.Require(a.preview))
	m.HandleFunc("GET /api/slices/{id}/archive", h.Require(a.sliceArchive))
	m.HandleFunc("GET /api/printers", h.Require(a.printers))
	m.HandleFunc("POST /api/printers", h.Require(a.addPrinter))
	m.HandleFunc("PATCH /api/printers/{id}", h.Require(a.patchPrinter))
	m.HandleFunc("POST /api/printers/discover", h.Require(a.discover))
	m.HandleFunc("POST /api/printers/{id}/probe", h.Require(a.probe))
	m.HandleFunc("POST /api/printers/{id}/trust", h.Require(a.trustPrinter))
	m.HandleFunc("POST /api/printers/{id}/test", h.Require(a.testPrinter))
	m.HandleFunc("GET /api/prints", h.Require(a.prints))
	m.HandleFunc("POST /api/printers/{id}/prints", h.Require(a.createPrint))
	m.HandleFunc("GET /api/prints/{id}", h.Require(a.printRoute))
	for _, action := range []string{"pause", "resume", "cancel", "reconcile", "resolve"} {
		m.HandleFunc("POST /api/prints/{id}/"+action, h.Require(a.controlPrint))
	}
	m.HandleFunc("GET /api/drive-saves", h.Require(a.driveSaves))
	m.HandleFunc("POST /api/drive-saves/{id}", h.Require(a.driveSave))
	m.HandleFunc("GET /api/drive-saves/{id}/file", h.Require(a.driveFile))
	return a, nil
}
func uid(r *http.Request) string { return apphost.Current(r).User.ID }
func now() string                { return time.Now().UTC().Format(time.RFC3339Nano) }
func validID(s string) bool      { return regexp.MustCompile(`^[a-f0-9]{48}$`).MatchString(s) }
func encode(v any) string        { b, _ := json.Marshal(v); return string(b) }
func (a *App) problem(w http.ResponseWriter, e error) {
	if e == nil {
		return
	}
	apphost.Problem(w, 400, e.Error())
}
func (a *App) system(w http.ResponseWriter, r *http.Request) {
	apphost.JSON(w, 200, map[string]any{"engine": a.Engine.State(), "engineSource": engine.Pin.Source, "source": "https://github.com/oneworkspacestudio/my-slicer", "models": []string{"A1", "A1 mini"}, "limits": map[string]any{"uploadMiB": 100, "triangles": 1500000, "workers": 1}, "owner": apphost.Current(r).User.Role == "owner"})
}
func (a *App) installEngine(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var active int
	a.H.DB.QueryRow("SELECT COUNT(*) FROM slicer_slices WHERE state IN ('queued','processing')").Scan(&active)
	if active != 0 {
		apphost.Problem(w, 409, "Finish or cancel active slices before engine installation")
		return
	}
	a.Engine.Install(a.ctx)
	apphost.JSON(w, 202, a.Engine.State())
}
func (a *App) profiles(w http.ResponseWriter, r *http.Request) {
	_, p, ready := a.Engine.Ready()
	if !ready {
		apphost.Problem(w, 409, "Finish native engine setup before choosing profiles")
		return
	}
	m, e := p.Machine(r.URL.Query().Get("model"))
	if e != nil {
		a.problem(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]any{"machine": m, "process": p.Compatible("process", m.Name), "filament": p.Compatible("filament", m.Name), "defaultProcess": engine.Text(m.Values["default_print_profile"]), "defaultFilament": engine.Text(m.Values["default_filament_profile"])})
}
func (a *App) Run(ctx context.Context) error {
	a.mu.Lock()
	a.ctx = ctx
	a.mu.Unlock()
	a.Engine.Check(ctx)
	a.wg.Add(1)
	go func() { defer a.wg.Done(); a.monitor(ctx) }()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	defer func() {
		a.Engine.Wait()
		a.devicesMu.Lock()
		for _, d := range a.devices {
			d.mu.Lock()
			if d.Client != nil {
				d.Client.Close()
			}
			d.mu.Unlock()
		}
		a.devicesMu.Unlock()
		a.wg.Wait()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			a.runNextSlice(ctx)
		}
	}
}

var errMissing = errors.New("This item is unavailable for the current account")
