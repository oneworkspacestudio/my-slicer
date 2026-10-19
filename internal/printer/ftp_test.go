package printer

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeFTPS struct {
	listener net.Listener
	mu       sync.Mutex
	files    map[string][]byte
	commands []string
	badSize  bool
	badLogin bool
}

func ftpFixture(t *testing.T, protected bool) (*Client, *fakeFTPS) {
	t.Helper()
	https := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	cert := https.TLS.Certificates[0]
	https.Close()
	raw, e := net.Listen("tcp4", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	server := &fakeFTPS{listener: tls.NewListener(raw, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}), files: map[string][]byte{}}
	t.Cleanup(func() { server.listener.Close() })
	go func() {
		for {
			c, e := server.listener.Accept()
			if e != nil {
				return
			}
			go server.serve(c, cert)
		}
	}()
	hash := sha256.Sum256(cert.Certificate[0])
	port := raw.Addr().(*net.TCPAddr).Port
	c := New(Record{FTPSPin: hex.EncodeToString(hash[:]), ProtectedData: protected}, "fixture-code")
	c.ip = net.ParseIP("127.0.0.1")
	c.ftpsPort = port
	return c, server
}
func (s *fakeFTPS) serve(conn net.Conn, cert tls.Certificate) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	reader := bufio.NewReader(conn)
	reply := func(v string) { fmt.Fprint(conn, v+"\r\n") }
	reply("220 fixture")
	var data net.Listener
	var pending string
	for {
		line, e := reader.ReadString('\n')
		if e != nil {
			return
		}
		parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
		cmd := parts[0]
		arg := ""
		if len(parts) > 1 {
			arg = parts[1]
		}
		s.mu.Lock()
		s.commands = append(s.commands, cmd+" "+arg)
		s.mu.Unlock()
		switch cmd {
		case "USER":
			reply("331 password")
		case "PASS":
			if s.badLogin {
				reply("530 wrong access code")
			} else {
				reply("230 welcome")
			}
		case "FEAT":
			reply("211-Features\r\n EPSV\r\n UTF8\r\n211 End")
		case "PBSZ", "TYPE", "OPTS":
			reply("200 ok")
		case "PROT":
			pending = arg
			reply("200 protected")
		case "EPSV":
			data, _ = net.Listen("tcp4", "127.0.0.1:0")
			if pending == "P" {
				data = tls.NewListener(data, &tls.Config{Certificates: []tls.Certificate{cert}})
			}
			reply(fmt.Sprintf("229 Entering Extended Passive Mode (|||%d|)", data.Addr().(*net.TCPAddr).Port))
		case "LIST", "STOR":
			reply("150 opening")
			d, e := data.Accept()
			data.Close()
			if e != nil {
				return
			}
			d.SetDeadline(time.Now().Add(5 * time.Second))
			if cmd == "LIST" {
				fmt.Fprint(d, "-rw-r--r-- 1 user group 12 Jan 01 2026 example.3mf\r\n")
			} else {
				b, _ := io.ReadAll(d)
				s.mu.Lock()
				s.files[arg] = b
				s.mu.Unlock()
			}
			d.Close()
			reply("226 complete")
		case "SIZE":
			s.mu.Lock()
			size := len(s.files[arg])
			s.mu.Unlock()
			if s.badSize {
				size++
			}
			reply("213 " + strconv.Itoa(size))
		case "RNFR":
			pending = arg
			reply("350 ready")
		case "RNTO":
			s.mu.Lock()
			s.files[arg] = s.files[pending]
			delete(s.files, pending)
			s.mu.Unlock()
			reply("250 renamed")
		case "QUIT":
			reply("221 bye")
			return
		default:
			reply("200 ok")
		}
	}
}
func TestActualFTPSControlDataAndFinalize(t *testing.T) {
	for _, protected := range []bool{false, true} {
		t.Run(fmt.Sprint(protected), func(t *testing.T) {
			c, s := ftpFixture(t, protected)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if e := c.TestFiles(ctx); e != nil {
				t.Fatal(e)
			}
			path := filepath.Join(t.TempDir(), "job.3mf")
			body := []byte(strings.Repeat("slice", 1000))
			os.WriteFile(path, body, 0600)
			var progress int64
			if e := c.Upload(ctx, path, "job.3mf", func(n int64) { progress = n }); e != nil {
				t.Fatal(e)
			}
			s.mu.Lock()
			defer s.mu.Unlock()
			if string(s.files["job.3mf"]) != string(body) || s.files["job.3mf.part"] != nil || progress != int64(len(body)) {
				t.Fatal("transfer not verified/finalized")
			}
			want := "PROT C"
			if protected {
				want = "PROT P"
			}
			if !strings.Contains(strings.Join(s.commands, "\n"), want) {
				t.Fatal("incorrect transfer protection")
			}
		})
	}
}
func TestFTPSRejectsWrongCodeAndIncompleteSize(t *testing.T) {
	c, s := ftpFixture(t, false)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s.badLogin = true
	if e := c.TestFiles(ctx); e == nil || !strings.Contains(e.Error(), "login failed") {
		t.Fatal(e)
	}
	s.badLogin = false
	s.badSize = true
	path := filepath.Join(t.TempDir(), "job.3mf")
	os.WriteFile(path, []byte(strings.Repeat("x", 2000)), 0600)
	if e := c.Upload(ctx, path, "incomplete.3mf", func(int64) {}); e == nil {
		t.Fatal("size mismatch accepted")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.files["incomplete.3mf"] != nil {
		t.Fatal("incomplete transfer finalized")
	}
}
