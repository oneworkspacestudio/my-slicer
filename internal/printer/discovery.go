package printer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type Discovery struct {
	Host   string `json:"host"`
	Name   string `json:"name"`
	Serial string `json:"serial"`
	Model  string `json:"model"`
}

func discover(ctx context.Context, target string, timeout time.Duration) ([]Discovery, error) {
	c, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if e != nil {
		return nil, e
	}
	defer c.Close()
	deadline := time.Now().Add(timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = c.SetDeadline(deadline)
	ip := net.ParseIP(target)
	if ip == nil {
		return nil, errors.New("Invalid discovery address")
	}
	message := fmt.Sprintf("M-SEARCH * HTTP/1.1\r\nHOST: %s:2021\r\nMAN: \"ssdp:discover\"\r\nMX: 1\r\nST: urn:bambulab-com:device:3dprinter:1\r\n\r\n", target)
	if _, e = c.WriteToUDP([]byte(message), &net.UDPAddr{IP: ip, Port: 2021}); e != nil {
		return nil, e
	}
	var out []Discovery
	seen := map[string]bool{}
	buf := make([]byte, 8192)
	for {
		n, addr, e := c.ReadFromUDP(buf)
		if e != nil {
			break
		}
		if !LAN(addr.IP) {
			continue
		}
		if target != "239.255.255.250" && !addr.IP.Equal(ip) {
			continue
		}
		headers := map[string]string{}
		for _, line := range strings.Split(string(buf[:n]), "\n") {
			parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
			if len(parts) == 2 {
				headers[strings.ToLower(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
		serial := strings.TrimPrefix(headers["usn"], "uuid:")
		if serial == "" || len(serial) > 40 || seen[serial] {
			continue
		}
		model := ModelName(headers["devmodel.bambu.com"])
		if model == "" {
			continue
		}
		name := headers["devname.bambu.com"]
		if name == "" {
			name = "Bambu Lab " + model
		}
		out = append(out, Discovery{Host: addr.IP.String(), Name: name, Serial: serial, Model: model})
		seen[serial] = true
	}
	return out, nil
}
func Discover(ctx context.Context) ([]Discovery, error) {
	return discover(ctx, "239.255.255.250", 3*time.Second)
}
func Observe(ctx context.Context, ip net.IP) (Discovery, error) {
	devices, e := discover(ctx, ip.String(), 2*time.Second)
	if e != nil {
		return Discovery{}, e
	}
	if len(devices) != 1 {
		return Discovery{}, errors.New("This printer did not advertise its identity; confirm the physical model/serial if discovery is blocked")
	}
	return devices[0], nil
}
