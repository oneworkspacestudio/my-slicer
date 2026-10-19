package printer

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func fixture() (Record, State, PrintSpec, PrintOptions) {
	r := Record{Model: "A1", DeveloperConfirmed: true}
	s := State{Connected: true, Seen: time.Now(), Status: "IDLE", IdentityVerified: true, Trays: []Tray{{Key: "0/3", Label: "AMS lite slot4", Wire: 3, Material: "PLA", NativeID: "GFA00", Color: "00AE42FF", Available: true}}}
	spec := PrintSpec{Remote: "my-slicer-" + strings.Repeat("a", 48) + ".gcode.3mf", GCode: "Metadata/plate_1.gcode", MD5: strings.Repeat("b", 32), TaskID: "12345", ModelID: "N2S", FilamentSlots: 1, Materials: []RequiredMaterial{{ID: 1, Type: "PLA", Color: "#00AE42", NativeID: "GFA00"}}}
	return r, s, spec, PrintOptions{UseAMS: true, Mapping: map[string]string{"1": "0/3"}, BedLeveling: true}
}
func TestMappingUsesActualSlotsAndNoUniversalSpoolSentinel(t *testing.T) {
	r, s, spec, o := fixture()
	payload, e := BuildStart(r, s, spec, o)
	if e != nil {
		t.Fatal(e)
	}
	p := payload["print"].(map[string]any)
	m := p["ams_mapping"].([]int)
	if len(m) != 1 || m[0] != 3 || p["md5"] != spec.MD5 || p["param"] != spec.GCode {
		t.Fatal(p)
	}
	spec.FilamentSlots = 3
	spec.Materials[0].ID = 3
	o.Mapping = map[string]string{"3": "0/3"}
	payload, e = BuildStart(r, s, spec, o)
	if e != nil {
		t.Fatal(e)
	}
	m = payload["print"].(map[string]any)["ams_mapping"].([]int)
	if len(m) != 3 || m[0] != -1 || m[1] != -1 || m[2] != 3 {
		t.Fatal(m)
	}
	o.UseAMS = false
	o.ExternalConfirmed = true
	payload, e = BuildStart(r, s, spec, o)
	if e != nil {
		t.Fatal(e)
	}
	if _, ok := payload["print"].(map[string]any)["ams_mapping"]; ok {
		t.Fatal("External spool emitted a guessed mapping sentinel")
	}
}
func TestDispatchRejectsUnsafeRealDeviceStates(t *testing.T) {
	r, s, spec, o := fixture()
	cases := []func(){func() { s.Status = "RUNNING" }, func() { s.Status = "IDLE"; s.IdentityVerified = false }, func() { s.IdentityVerified = true; spec.ModelID = "N1" }, func() { spec.ModelID = "N2S"; s.Trays[0].Available = false }, func() { s.Trays[0].Available = true; s.Trays[0].Material = "PETG" }}
	for _, change := range cases {
		change()
		if _, e := BuildStart(r, s, spec, o); e == nil {
			t.Fatal("Unsafe dispatch accepted")
		}
	}
}
func TestAcknowledgementIsNotAStatusAndHexModeIsCorrect(t *testing.T) {
	c := New(Record{Model: "A1", ManualIdentity: true}, "disposable-code")
	c.receive([]byte(`{"print":{"command":"project_file","sequence_id":"7","result":"success","gcode_state":"RUNNING"}}`))
	if c.State().Status != "UNKNOWN" {
		t.Fatal("Echo fabricated physical printing")
	}
	if _, ok := c.Ack("7"); !ok {
		t.Fatal("Real acknowledgement lost")
	}
	c.receive([]byte(`{"fun":"20000000","print":{"command":"push_status","gcode_state":"IDLE","mc_percent":0}}`))
	s := c.State()
	if s.DeveloperMode == nil || *s.DeveloperMode {
		t.Fatal("Verification-required hexadecimal flag was read as decimal")
	}
	c.receive([]byte(`{"print":{"command":"push_status","fun":"0","gcode_state":"IDLE"}}`))
	if mode := c.State().DeveloperMode; mode == nil || !*mode {
		t.Fatal("Developer mode flag not updated")
	}
}
func TestPrivateNetworkAndDeviceScopedTLS(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "8.8.8.8", "169.254.169.254", "0.0.0.0", "239.255.255.250"} {
		if LAN(net.ParseIP(ip)) {
			t.Fatal("Unsafe destination", ip)
		}
	}
	if !LAN(net.ParseIP("192.168.1.9")) {
		t.Fatal("LAN rejected")
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) }))
	defer server.Close()
	cert := server.Certificate()
	address := strings.TrimPrefix(server.URL, "https://")
	conn, e := tls.Dial("tcp", address, TLS(Fingerprint(cert.Raw), "localhost"))
	if e != nil {
		t.Fatal(e)
	}
	conn.Close()
	if conn, e = tls.Dial("tcp", address, TLS(strings.Repeat("0", 64), "localhost")); e == nil {
		conn.Close()
		t.Fatal("Mismatched pinned certificate accepted")
	}
}
func TestAccessCodeProtectedForCurrentWindowsUser(t *testing.T) {
	cipher, e := Protect("fixture-code-1234")
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(cipher), "fixture-code-1234") {
		t.Fatal("Clear credential persisted")
	}
	code, e := Unprotect(cipher)
	if e != nil || code != "fixture-code-1234" {
		t.Fatal(e)
	}
}
