package printer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Ack struct {
	Result string
	Reason string
	At     time.Time
}
type Client struct {
	ftpsPort      int // package-local protocol fixture override; production always uses 990
	Record        Record
	code          string
	ip            net.IP
	mu            sync.RWMutex
	state         State
	raw           map[string]any
	acks          map[string]Ack
	client        mqtt.Client
	observed      bool
	identityError bool
}

func New(r Record, code string) *Client {
	return &Client{Record: r, code: code, raw: map[string]any{}, acks: map[string]Ack{}, state: State{Status: "UNKNOWN", Trays: []Tray{}}}
}
func (c *Client) State() State {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s := c.state
	s.Trays = append([]Tray{}, s.Trays...)
	if !s.Fresh() {
		s.Connected = false
	}
	return s
}
func (c *Client) Ack(sequence string) (Ack, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.acks[sequence]
	return v, ok
}
func (c *Client) Close() {
	c.mu.Lock()
	client := c.client
	c.client = nil
	c.state.Connected = false
	c.mu.Unlock()
	if client != nil {
		client.Disconnect(200)
	}
}
func (c *Client) Connect(ctx context.Context) error {
	ip, e := Resolve(ctx, c.Record.Host)
	if e != nil {
		return e
	}
	c.ip = ip
	if observed, e := Observe(ctx, ip); e == nil {
		if observed.Serial != c.Record.Serial || observed.Model != c.Record.Model {
			return errors.New("Printer identity/model does not match the saved device; verify the IP and serial before pairing again")
		}
		c.mu.Lock()
		c.observed = true
		c.state.ObservedModel = observed.Model
		c.state.IdentityVia = "LAN identity advertisement and authenticated serial-specific status"
		c.mu.Unlock()
	}
	if len(c.Record.MQTTPin) != 64 || len(c.Record.FTPSPin) != 64 {
		return errors.New("Review and trust this printer's two local certificates first")
	}
	opts := mqtt.NewClientOptions().AddBroker("ssl://" + net.JoinHostPort(ip.String(), "8883")).SetClientID("myslicer-" + strconv.FormatInt(time.Now().UnixNano(), 36)).SetUsername("bblp").SetPassword(c.code).SetTLSConfig(TLS(c.Record.MQTTPin, ip.String())).SetCleanSession(true).SetAutoReconnect(false).SetConnectRetry(false).SetKeepAlive(15 * time.Second).SetPingTimeout(8 * time.Second).SetWriteTimeout(10 * time.Second).SetConnectTimeout(10 * time.Second)
	opts.SetConnectionLostHandler(func(_ mqtt.Client, _ error) {
		c.mu.Lock()
		c.state.Connected = false
		c.state.Error = "Local status connection ended; reconnect to reconcile any running print"
		c.mu.Unlock()
	})
	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(12 * time.Second) {
		client.Disconnect(0)
		return errors.New("Printer MQTT connection timed out; check LAN mode, IP and firewall")
	}
	if token.Error() != nil {
		client.Disconnect(0)
		return errors.New("Printer status authentication/TLS failed; check the LAN access code and certificate trust")
	}
	c.mu.Lock()
	c.client = client
	c.state.Connected = true
	c.mu.Unlock()
	token = client.Subscribe("device/"+c.Record.Serial+"/report", 0, func(_ mqtt.Client, msg mqtt.Message) {
		if len(msg.Payload()) > 2<<20 {
			return
		}
		c.receive(msg.Payload())
	})
	if !token.WaitTimeout(10*time.Second) || token.Error() != nil {
		c.Close()
		return errors.New("Printer status subscription failed; check its serial and LAN permissions")
	}
	_ = c.Refresh()
	deadline := time.NewTimer(12 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			c.Close()
			return ctx.Err()
		case <-deadline.C:
			if c.State().Fresh() {
				c.Close()
				return errors.New("Authenticated status arrived but printer identity is unconfirmed. Verify serial/model on the physical screen and explicitly confirm that identity in setup")
			}
			c.Close()
			return errors.New("Authenticated but no serial-specific printer status arrived; verify the serial and LAN/Developer mode")
		case <-tick.C:
			s := c.State()
			if s.Fresh() {
				if s.IdentityVerified {
					return nil
				}
			}
		}
	}
}
func (c *Client) send(payload any) error {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	if client == nil || !client.IsConnectionOpen() {
		return errors.New("Printer is disconnected; no command was sent")
	}
	b, e := json.Marshal(payload)
	if e != nil {
		return e
	}
	token := client.Publish("device/"+c.Record.Serial+"/request", 0, false, b)
	if !token.WaitTimeout(12 * time.Second) {
		return errors.New("Printer command transport timed out; reconcile device state before any retry")
	}
	return token.Error()
}
func (c *Client) Refresh() error {
	if e := c.send(map[string]any{"pushing": map[string]any{"command": "pushall", "sequence_id": "0"}}); e != nil {
		return e
	}
	return c.send(map[string]any{"info": map[string]any{"command": "get_version", "sequence_id": "0"}})
}
func number(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case string:
		n, e := strconv.ParseFloat(x, 64)
		return n, e == nil
	}
	return 0, false
}
func text(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
	return ""
}
func merge(dst, src map[string]any) {
	for k, v := range src {
		if sub, ok := v.(map[string]any); ok {
			if old, ok := dst[k].(map[string]any); ok {
				merge(old, sub)
				continue
			}
		}
		dst[k] = v
	}
}
func (c *Client) receive(b []byte) {
	var packet map[string]any
	if json.Unmarshal(b, &packet) != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if info, ok := packet["info"].(map[string]any); ok {
		modules, _ := info["module"].([]any)
		for _, v := range modules {
			m, ok := v.(map[string]any)
			if !ok {
				continue
			}
			if text(m["name"]) == "ota" {
				c.state.Firmware = text(m["sw_ver"])
			}
			if observed := ModelName(text(m["project_name"])); observed != "" {
				c.state.ObservedModel = observed
				if observed != c.Record.Model {
					c.identityError = true
					c.state.Error = "The connected device reports a different printer model"
				}
			}
		}
	}
	status, ok := packet["print"].(map[string]any)
	if !ok {
		return
	}
	command := text(status["command"])
	if command != "" && command != "push_status" {
		if sequence := text(status["sequence_id"]); sequence != "" {
			reason := strings.ReplaceAll(text(status["reason"]), c.code, "[redacted]")
			if len(reason) > 200 {
				reason = reason[:200]
			}
			c.acks[sequence] = Ack{Result: strings.ToLower(text(status["result"])), Reason: reason, At: time.Now().UTC()}
			if len(c.acks) > 100 {
				for k, v := range c.acks {
					if time.Since(v.At) > 5*time.Minute {
						delete(c.acks, k)
					}
				}
			}
		}
		return // A project_file echo/ack is not a physical printer status frame.
	}
	merge(c.raw, status)
	c.state.Seen = time.Now().UTC()
	c.state.Connected = true
	c.state.Status = strings.ToUpper(text(c.raw["gcode_state"]))
	if c.state.Status == "" {
		c.state.Status = "UNKNOWN"
	}
	fun, hasFun := status["fun"]
	if !hasFun {
		fun, hasFun = packet["fun"]
	}
	if hasFun {
		var bits uint64
		valid := false
		switch x := fun.(type) {
		case string:
			v, e := strconv.ParseUint(strings.TrimPrefix(x, "0x"), 16, 64)
			bits = v
			valid = e == nil
		case float64:
			bits = uint64(x)
			valid = true
		}
		if valid {
			mode := bits&0x20000000 == 0
			c.state.DeveloperMode = &mode
		}
	}
	setInt := func(key string, dst **int) {
		if n, ok := number(c.raw[key]); ok && n >= 0 && n <= 100000 {
			v := int(n)
			*dst = &v
		}
	}
	setFloat := func(key string, dst **float64) {
		if n, ok := number(c.raw[key]); ok && n >= 0 && n <= 1000 {
			v := n
			*dst = &v
		}
	}
	setInt("mc_percent", &c.state.Progress)
	setInt("mc_remaining_time", &c.state.Remaining)
	setInt("layer_num", &c.state.Layer)
	setInt("total_layer_num", &c.state.TotalLayers)
	setFloat("nozzle_temper", &c.state.Nozzle)
	setFloat("bed_temper", &c.state.Bed)
	if sd, ok := c.raw["sdcard"]; ok {
		value := false
		switch v := sd.(type) {
		case bool:
			value = v
		case float64:
			value = v != 0
		case string:
			value = strings.Contains(strings.ToUpper(v), "HAS_SDCARD") || v == "1" || strings.EqualFold(v, "normal") || strings.EqualFold(v, "true")
		}
		c.state.SDCard = &value
	}
	c.state.File = text(c.raw["gcode_file"])
	c.state.Subtask = text(c.raw["subtask_name"])
	c.state.TaskID = text(c.raw["subtask_id"])
	if v := text(c.raw["print_error"]); v != "" && v != "0" {
		c.state.PrintError = "Printer reported error code " + v
	} else {
		c.state.PrintError = ""
	}
	c.state.Trays = parseTrays(c.raw)
	c.state.IdentityVerified = !c.identityError && (c.observed || c.state.ObservedModel == c.Record.Model || c.Record.ManualIdentity)
	if c.Record.ManualIdentity && !c.observed && c.state.ObservedModel == "" {
		c.state.IdentityVia = "Physical-screen confirmation and authenticated exact-serial telemetry"
	}
	if c.state.IdentityVerified {
		c.state.Error = ""
	}
}
func parseTrays(raw map[string]any) []Tray {
	out := []Tray{}
	ams, ok := raw["ams"].(map[string]any)
	if !ok {
		return out
	}
	units, _ := ams["ams"].([]any)
	for _, v := range units {
		unit, ok := v.(map[string]any)
		if !ok {
			continue
		}
		id, e := strconv.Atoi(text(unit["id"]))
		if e != nil || id != 0 {
			continue
		}
		trays, _ := unit["tray"].([]any)
		for _, v := range trays {
			tray, ok := v.(map[string]any)
			if !ok {
				continue
			}
			slot, e := strconv.Atoi(text(tray["id"]))
			if e != nil || slot < 0 || slot > 3 {
				continue
			}
			material := text(tray["tray_type"])
			available := material != "" && text(tray["tray_info_idx"]) != ""
			if mask, exists := ams["tray_exist_bits"]; exists {
				bits, e := strconv.ParseUint(strings.TrimPrefix(text(mask), "0x"), 16, 64)
				if e == nil {
					available = available && bits&(1<<uint(slot)) != 0
				}
			}
			out = append(out, Tray{Key: fmt.Sprintf("%d/%d", id, slot), Label: fmt.Sprintf("AMS lite · slot %d", slot+1), Unit: id, Slot: slot, Wire: slot, Material: material, Color: text(tray["tray_color"]), NativeID: text(tray["tray_info_idx"]), Available: available})
		}
	}
	return out
}
