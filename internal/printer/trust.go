package printer

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"
)

func Resolve(ctx context.Context, host string) (net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		if !LAN(ip) {
			return nil, errors.New("Only private IPv4 LAN printer addresses are allowed")
		}
		return ip, nil
	}
	addresses, e := net.DefaultResolver.LookupIPAddr(ctx, host)
	if e != nil {
		return nil, errors.New("Printer hostname could not be resolved; check its current LAN IP")
	}
	if len(addresses) == 0 {
		return nil, errors.New("Printer hostname has no address")
	}
	for _, a := range addresses {
		if !LAN(a.IP) {
			return nil, errors.New("Printer hostnames must resolve only to private IPv4 LAN addresses")
		}
	}
	return addresses[0].IP, nil
}
func Fingerprint(cert []byte) string { b := sha256.Sum256(cert); return hex.EncodeToString(b[:]) }
func TLS(pin, host string) *tls.Config {
	// Device-scoped explicit certificate pinning handles Bambu's self-signed LAN
	// certificate. No global TLS client or unrelated outbound request is weakened.
	return &tls.Config{MinVersion: tls.VersionTLS12, ServerName: host, InsecureSkipVerify: true, ClientSessionCache: tls.NewLRUClientSessionCache(8), VerifyConnection: func(s tls.ConnectionState) error {
		if len(s.PeerCertificates) == 0 || len(pin) != 64 || subtle.ConstantTimeCompare([]byte(Fingerprint(s.PeerCertificates[0].Raw)), []byte(strings.ToLower(pin))) != 1 {
			return errors.New("Printer certificate changed or has not been explicitly trusted; pair this device again")
		}
		return nil
	}}
}
func ProbeCertificate(ctx context.Context, ip net.IP, port int) (Certificate, error) {
	var out Certificate
	d := net.Dialer{Timeout: 8 * time.Second}
	c, e := d.DialContext(ctx, "tcp", net.JoinHostPort(ip.String(), strconv.Itoa(port)))
	if e != nil {
		return out, errors.New("Printer is unreachable on its local secure service; check LAN mode, IP and firewall")
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(8 * time.Second))
	// This handshake only reads an untrusted certificate for display. No password
	// is sent and it is not persisted as trusted until the user confirms pairing.
	t := tls.Client(c, &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true})
	if e = t.HandshakeContext(ctx); e != nil {
		return out, errors.New("The printer did not complete its local TLS handshake")
	}
	s := t.ConnectionState()
	if len(s.PeerCertificates) == 0 {
		return out, errors.New("The printer supplied no TLS certificate")
	}
	cert := s.PeerCertificates[0]
	out = Certificate{SHA256: Fingerprint(cert.Raw), Subject: cert.Subject.CommonName, Expires: cert.NotAfter}
	return out, nil
}
