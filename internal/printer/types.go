package printer

import (
	"errors"
	"net"
	"regexp"
	"strings"
	"time"
)

type Record struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Model              string `json:"model"`
	Host               string `json:"host"`
	Serial             string `json:"serial"`
	Nozzle             string `json:"nozzle"`
	Plate              string `json:"plate"`
	Revision           int    `json:"revision"`
	MQTTPin            string `json:"mqttPin"`
	FTPSPin            string `json:"ftpsPin"`
	ProtectedData      bool   `json:"protectedData"`
	ManualIdentity     bool   `json:"manualIdentity"`
	DeveloperConfirmed bool   `json:"developerConfirmed"`
}
type Tray struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Unit      int    `json:"unit"`
	Slot      int    `json:"slot"`
	Wire      int    `json:"wire"`
	Material  string `json:"material"`
	Color     string `json:"color"`
	NativeID  string `json:"nativeID"`
	Available bool   `json:"available"`
}
type State struct {
	Connected        bool      `json:"connected"`
	IdentityVerified bool      `json:"identityVerified"`
	IdentityVia      string    `json:"identityVia"`
	ObservedModel    string    `json:"observedModel"`
	Firmware         string    `json:"firmware"`
	Status           string    `json:"status"`
	Seen             time.Time `json:"seen"`
	Progress         *int      `json:"progress"`
	Remaining        *int      `json:"remaining"`
	Layer            *int      `json:"layer"`
	TotalLayers      *int      `json:"totalLayers"`
	Nozzle           *float64  `json:"nozzleTemperature"`
	Bed              *float64  `json:"bedTemperature"`
	SDCard           *bool     `json:"sdcard"`
	DeveloperMode    *bool     `json:"developerMode"`
	Trays            []Tray    `json:"trays"`
	File             string    `json:"file"`
	Subtask          string    `json:"subtask"`
	TaskID           string    `json:"taskID"`
	Error            string    `json:"error"`
	PrintError       string    `json:"printError"`
}
type Diagnostic struct {
	Stage   string `json:"stage"`
	State   string `json:"state"`
	Message string `json:"message"`
}
type Certificate struct {
	SHA256  string    `json:"sha256"`
	Subject string    `json:"subject"`
	Expires time.Time `json:"expires"`
}
type Pairing struct {
	MQTT   Certificate `json:"mqtt"`
	FTPS   Certificate `json:"ftps"`
	Model  string      `json:"observedModel"`
	Serial string      `json:"observedSerial"`
	Name   string      `json:"observedName"`
}

func (r Record) Validate() error {
	if strings.TrimSpace(r.Name) == "" || len(r.Name) > 100 {
		return errors.New("Give this printer a name of up to100 characters")
	}
	if r.Model != "A1" && r.Model != "A1 mini" {
		return errors.New("Choose A1 or A1 mini")
	}
	if r.Nozzle != "0.4" {
		return errors.New("Confirm that a physical0.4mm nozzle is installed")
	}
	if !regexp.MustCompile(`^[A-Za-z0-9]{8,40}$`).MatchString(r.Serial) {
		return errors.New("Enter the serial number shown on this printer")
	}
	if r.Plate != "Textured PEI Plate" && r.Plate != "High Temp Plate" {
		return errors.New("Choose the confirmed Textured PEI or Smooth PEI/High Temp plate")
	}
	if strings.ContainsAny(r.Host, "/\\:@ \r\n") || len(r.Host) > 253 || r.Host == "" {
		return errors.New("Enter a local printer IP or hostname, without a URL or port")
	}
	return nil
}
func LAN(ip net.IP) bool {
	return ip != nil && ip.To4() != nil && ip.IsPrivate() && !ip.IsLoopback() && !ip.IsUnspecified() && !ip.IsMulticast()
}
func ModelName(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "N2S", "A1", "BAMBU LAB A1":
		return "A1"
	case "N1", "A1 MINI", "BAMBU LAB A1 MINI":
		return "A1 mini"
	}
	return ""
}
func (s State) Fresh() bool {
	return s.Connected && !s.Seen.IsZero() && time.Since(s.Seen) < 30*time.Second
}
func (s State) Idle() bool {
	return s.Fresh() && (s.Status == "IDLE" || s.Status == "FINISH" || s.Status == "FAILED")
}
func (s State) CommandMode(r Record) bool {
	return s.DeveloperMode != nil && *s.DeveloperMode || s.DeveloperMode == nil && r.DeveloperConfirmed
}
