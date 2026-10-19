package engine

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:embed runtime-manifest.json
var manifestBytes []byte

type RuntimeFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}
type Artifact struct {
	Version      string        `json:"version"`
	Tag          string        `json:"tag"`
	URL          string        `json:"url"`
	Size         int64         `json:"size"`
	SHA256       string        `json:"sha256"`
	Source       string        `json:"source"`
	SourceCommit string        `json:"sourceCommit"`
	Files        []RuntimeFile `json:"files"`
}

var Pin = func() Artifact {
	var a Artifact
	if e := json.Unmarshal(manifestBytes, &a); e != nil {
		panic(e)
	}
	return a
}()

type InstallState struct {
	Phase   string    `json:"phase"`
	Message string    `json:"message"`
	Done    int64     `json:"done"`
	Total   int64     `json:"total"`
	Unit    string    `json:"unit"`
	Updated time.Time `json:"updated"`
	Version string    `json:"version"`
	Busy    bool      `json:"busy"`
}
type Manager struct {
	FileLock *sync.RWMutex
	base     string
	mu       sync.RWMutex
	state    InstallState
	root     string
	profiles *Profiles
	cancel   context.CancelFunc
	running  bool
	wg       sync.WaitGroup
}

func NewManager(data string) *Manager {
	return &Manager{base: filepath.Join(data, "engine"), state: InstallState{Phase: "not_installed", Message: "Install the native slicing engine to prepare A1 and A1 mini profiles", Version: Pin.Version}}
}
func (m *Manager) State() InstallState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s := m.state
	s.Busy = m.running
	return s
}
func (m *Manager) set(phase, message, unit string, done, total int64) {
	m.mu.Lock()
	m.state = InstallState{Phase: phase, Message: message, Done: done, Total: total, Unit: unit, Updated: time.Now().UTC(), Version: Pin.Version}
	m.mu.Unlock()
}
func (m *Manager) Ready() (string, *Profiles, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.root, m.profiles, m.root != "" && m.profiles != nil && !m.running
}
func (m *Manager) Cancel() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
	}
}
func (m *Manager) Wait() { m.Cancel(); m.wg.Wait() }
func (m *Manager) Check(ctx context.Context) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.wg.Add(1)
	m.mu.Unlock()
	go func() {
		defer m.wg.Done()
		if m.FileLock != nil {
			m.FileLock.RLock()
			defer m.FileLock.RUnlock()
		}
		defer func() { m.mu.Lock(); m.running = false; m.mu.Unlock() }()
		var current struct {
			Version string `json:"version"`
		}
		b, e := os.ReadFile(filepath.Join(m.base, "current.json"))
		if os.IsNotExist(e) {
			return
		}
		if e != nil || json.Unmarshal(b, &current) != nil || current.Version != Pin.Version {
			m.set("update_required", "Install this app's pinned engine version; the previous installation is retained", "", 0, 0)
			return
		}
		root := filepath.Join(m.base, Pin.Version)
		if e = m.verify(ctx, root); e != nil {
			m.set("error", "The saved engine could not be verified. Reinstall it before slicing", "", 0, 0)
			return
		}
		p, e := LoadProfiles(filepath.Join(root, "resources", "profiles", "BBL"))
		if e == nil {
			e = PrepareResources(root, p)
		}
		if e != nil {
			m.set("error", "Release-matched printer profiles could not be resolved: "+e.Error(), "", 0, 0)
			return
		}
		m.mu.Lock()
		m.root = root
		m.profiles = p
		m.mu.Unlock()
		m.set("ready", "Native Bambu Studio engine and A1/A1 mini profiles verified", "", 0, 0)
	}()
}
func (m *Manager) Install(ctx context.Context) bool {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return false
	}
	m.running = true
	ctx, m.cancel = context.WithTimeout(ctx, 30*time.Minute)
	m.wg.Add(1)
	m.mu.Unlock()
	go func() {
		defer m.wg.Done()
		if m.FileLock != nil {
			m.FileLock.RLock()
			defer m.FileLock.RUnlock()
		}
		defer func() {
			m.mu.Lock()
			if m.cancel != nil {
				m.cancel()
			}
			m.cancel = nil
			m.running = false
			m.mu.Unlock()
		}()
		if e := m.install(ctx); e != nil {
			message := e.Error()
			if ctx.Err() != nil {
				message = "Engine setup was cancelled or timed out. The previous working engine is retained; retry setup when ready"
			}
			m.set("error", message, "", 0, 0)
		}
	}()
	return true
}
func (m *Manager) verify(ctx context.Context, root string) error {
	m.set("verifying", "Checking the installed native engine and release resources", "files", 0, int64(len(Pin.Files)))
	for i, f := range Pin.Files {
		if e := ctx.Err(); e != nil {
			return e
		}
		p := filepath.Join(root, filepath.FromSlash(f.Path))
		st, e := os.Lstat(p)
		if e != nil || !st.Mode().IsRegular() || st.Size() != f.Size {
			return errors.New("Engine file mismatch")
		}
		if e = checkHash(p, f.Size, f.SHA256); e != nil {
			return e
		}
		if i%50 == 0 {
			m.set("verifying", "Checking native engine files and resources", "files", int64(i+1), int64(len(Pin.Files)))
		}
	}
	return nil
}
func checkHash(path string, size int64, sum string) error {
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	h := sha256.New()
	n, e := io.Copy(h, io.LimitReader(f, size+1))
	if e != nil || n != size || hex.EncodeToString(h.Sum(nil)) != sum {
		return errors.New("Official engine checksum/size did not match; no unverified code was executed")
	}
	return nil
}
func (m *Manager) install(ctx context.Context) error {
	if e := os.MkdirAll(m.base, 0700); e != nil {
		return e
	}
	stage, e := os.MkdirTemp(m.base, ".install-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(stage)
	archive := os.Getenv("MYSLICER_ENGINE_ARCHIVE")
	if archive == "" {
		archive = filepath.Join(stage, "engine.zip")
		if e = m.download(ctx, archive); e != nil {
			return e
		}
	} else {
		m.set("verifying", "Checking the supplied offline official engine archive", "bytes", 0, Pin.Size)
	}
	m.set("verifying", "Checking the complete official archive checksum before extraction", "", 0, 0)
	if e = checkHash(archive, Pin.Size, Pin.SHA256); e != nil {
		return e
	}
	runtime := filepath.Join(stage, "runtime")
	if e = m.extract(ctx, archive, runtime); e != nil {
		return e
	}
	m.set("testing", "Running real native slicing checks for A1 and A1 mini", "", 0, 0)
	if e = Validate(ctx, runtime, filepath.Join(stage, "validation"), func(s string) { m.set("testing", s, "", 0, 0) }); e != nil {
		return fmt.Errorf("Engine acceptance failed; previous version retained: %w", e)
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	p, e := LoadProfiles(filepath.Join(runtime, "resources", "profiles", "BBL"))
	if e != nil {
		return e
	}
	pointer := filepath.Join(m.base, ".current-new.json")
	if e = WriteJSON(pointer, map[string]string{"version": Pin.Version, "archiveSHA256": Pin.SHA256, "sourceCommit": Pin.SourceCommit}); e != nil {
		return e
	}
	target := filepath.Join(m.base, Pin.Version)
	retired := ""
	if _, e = os.Lstat(target); e == nil {
		retired = target + "-retained-" + time.Now().UTC().Format("20060102T150405.000000000")
		if e = os.Rename(target, retired); e != nil {
			return errors.New("The previous engine is in use; finish slicing and retry setup")
		}
	}
	if e = os.Rename(runtime, target); e != nil {
		if retired != "" {
			if restore := os.Rename(retired, target); restore != nil {
				return fmt.Errorf("Engine promotion failed and the retained version needs recovery: %v", restore)
			}
		}
		return e
	}
	if e = os.Rename(pointer, filepath.Join(m.base, "current.json")); e != nil {
		if restore := os.Rename(target, runtime); restore != nil {
			return fmt.Errorf("Engine pointer commit failed; recovery required: %v", restore)
		}
		if retired != "" {
			if restore := os.Rename(retired, target); restore != nil {
				return fmt.Errorf("Engine pointer commit failed; retained version needs recovery: %v", restore)
			}
		}
		return e
	}
	for key, entry := range p.Entries {
		entry.File = strings.Replace(entry.File, runtime, target, 1)
		p.Entries[key] = entry
	}
	m.mu.Lock()
	m.root = target
	m.profiles = p
	m.mu.Unlock()
	m.set("ready", "Real A1 and A1 mini native slicing checks passed; engine ready", "", 0, 0)
	return nil
}
func (m *Manager) download(ctx context.Context, path string) error {
	m.set("downloading", "Downloading the official pinned Bambu Studio Windows runtime", "bytes", 0, Pin.Size)
	client := &http.Client{Timeout: 25 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 4 {
			return errors.New("Too many engine download redirects")
		}
		if !officialURL(req.URL) {
			return errors.New("Engine download redirected outside official release storage")
		}
		return nil
	}}
	req, e := http.NewRequestWithContext(ctx, "GET", Pin.URL, nil)
	if e != nil {
		return e
	}
	req.Header.Set("User-Agent", "OneWorkspace-MySlicer/0.1")
	resp, e := client.Do(req)
	if e != nil {
		return errors.New("Engine download could not connect. Check Internet access and retry")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 || resp.ContentLength > 0 && resp.ContentLength != Pin.Size {
		return errors.New("Official engine download returned an unexpected response; retry setup")
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	buf := make([]byte, 128<<10)
	var done int64
	last := time.Time{}
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			done += int64(n)
			if done > Pin.Size {
				return errors.New("Engine download exceeded the pinned size")
			}
			if _, e = f.Write(buf[:n]); e != nil {
				return e
			}
			if time.Since(last) > 250*time.Millisecond {
				m.set("downloading", "Downloading official native engine; validation and printer-profile checks follow", "bytes", done, Pin.Size)
				last = time.Now()
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return errors.New("Engine download was interrupted; retry setup")
		}
	}
	if done != Pin.Size {
		return errors.New("Engine download was incomplete; retry setup")
	}
	return f.Sync()
}
func officialURL(u *url.URL) bool {
	return u.Scheme == "https" && u.User == nil && (u.Host == "github.com" || u.Host == "release-assets.githubusercontent.com" || u.Host == "objects.githubusercontent.com")
}
func (m *Manager) extract(ctx context.Context, archive, root string) error {
	z, e := zip.OpenReader(archive)
	if e != nil {
		return e
	}
	defer z.Close()
	members := map[string]*zip.File{}
	for _, f := range z.File {
		if members[f.Name] != nil {
			return errors.New("Duplicate native archive member")
		}
		members[f.Name] = f
	}
	var total int64
	for _, f := range Pin.Files {
		total += f.Size
	}
	var done int64
	for i, entry := range Pin.Files {
		if e = ctx.Err(); e != nil {
			return e
		}
		rel := filepath.FromSlash(entry.Path)
		if filepath.IsAbs(rel) || strings.Contains(entry.Path, "..") || strings.ContainsAny(entry.Path, "\\:") {
			return errors.New("Unsafe pinned runtime path")
		}
		f := members[entry.Path]
		if f == nil || f.UncompressedSize64 != uint64(entry.Size) || f.Mode()&os.ModeSymlink != 0 {
			return errors.New("Native archive is missing a required runtime file")
		}
		dest := filepath.Join(root, rel)
		if e = os.MkdirAll(filepath.Dir(dest), 0700); e != nil {
			return e
		}
		src, e := f.Open()
		if e != nil {
			return e
		}
		dst, e := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			src.Close()
			return e
		}
		h := sha256.New()
		n, e := io.Copy(io.MultiWriter(dst, h), io.LimitReader(src, entry.Size+1))
		src.Close()
		closeErr := dst.Close()
		if e != nil {
			return e
		}
		if closeErr != nil {
			return closeErr
		}
		if n != entry.Size || hex.EncodeToString(h.Sum(nil)) != entry.SHA256 {
			return errors.New("Native runtime extraction failed its checksum check")
		}
		done += n
		if i%20 == 0 {
			m.set("installing", "Installing native runtime and complete release-matched resources", "bytes", done, total)
		}
	}
	return nil
}
