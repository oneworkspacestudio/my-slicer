package printer

import (
	"context"
	"crypto/tls"
	"errors"
	ftp "github.com/jlaffaye/ftp"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type idleConn struct{ net.Conn }

func (c idleConn) Read(b []byte) (int, error) {
	_ = c.Conn.SetReadDeadline(time.Now().Add(35 * time.Second))
	return c.Conn.Read(b)
}
func (c idleConn) Write(b []byte) (int, error) {
	_ = c.Conn.SetWriteDeadline(time.Now().Add(35 * time.Second))
	return c.Conn.Write(b)
}
func (c *Client) ftp(ctx context.Context) (*ftp.ServerConn, func(), error) {
	if c.ip == nil {
		return nil, nil, errors.New("Connect and verify the printer identity first")
	}
	config := TLS(c.Record.FTPSPin, c.ip.String())
	controlPort := "990"
	if c.ftpsPort != 0 {
		controlPort = strconv.Itoa(c.ftpsPort)
	}
	var mu sync.Mutex
	var connections []net.Conn
	done := make(chan struct{})
	var once sync.Once
	closeAll := func() {
		once.Do(func() {
			close(done)
			mu.Lock()
			defer mu.Unlock()
			for _, conn := range connections {
				_ = conn.Close()
			}
		})
	}
	go func() {
		select {
		case <-ctx.Done():
			closeAll()
		case <-done:
		}
	}()
	dial := func(network, address string) (net.Conn, error) {
		host, port, e := net.SplitHostPort(address)
		if e != nil || !net.ParseIP(host).Equal(c.ip) {
			return nil, errors.New("Passive transfer redirected away from the paired printer")
		}
		n, e := strconv.Atoi(port)
		if e != nil || n < 1 || n > 65535 {
			return nil, errors.New("Invalid printer transfer port")
		}
		d := net.Dialer{Timeout: 10 * time.Second}
		raw, e := d.DialContext(ctx, "tcp", address)
		if e != nil {
			return nil, e
		}
		var conn net.Conn = idleConn{raw}
		if port == controlPort || c.Record.ProtectedData {
			t := tls.Client(conn, config)
			if port == controlPort {
				if e = t.HandshakeContext(ctx); e != nil {
					raw.Close()
					return nil, e
				}
			}
			conn = t
		}
		mu.Lock()
		select {
		case <-done:
			mu.Unlock()
			conn.Close()
			return nil, context.Canceled
		default:
		}
		connections = append(connections, conn)
		mu.Unlock()
		return conn, nil
	}
	connection, e := ftp.Dial(net.JoinHostPort(c.ip.String(), controlPort), ftp.DialWithTLS(config), ftp.DialWithClearData(!c.Record.ProtectedData), ftp.DialWithDialFunc(dial), ftp.DialWithTrustPasvIP(false), ftp.DialWithDisabledMLSD(true), ftp.DialWithTimeout(10*time.Second), ftp.DialWithShutTimeout(20*time.Second))
	if e != nil {
		closeAll()
		return nil, nil, errors.New("Printer FTPS connection failed; check LAN mode, storage and its trusted certificate")
	}
	if e = connection.Login("bblp", c.code); e != nil {
		connection.Quit()
		closeAll()
		return nil, nil, errors.New("Printer FTPS login failed; verify the LAN access code and local file-transfer permissions")
	}
	return connection, closeAll, nil
}
func (c *Client) TestFiles(ctx context.Context) error {
	connection, closeAll, e := c.ftp(ctx)
	if e != nil {
		return e
	}
	defer closeAll()
	defer connection.Quit()
	if _, e = connection.List("/"); e != nil {
		return errors.New("Authenticated FTPS but storage listing/data channel failed; check the SD card and this A1's transfer mode")
	}
	return nil
}

type countedReader struct {
	r      io.Reader
	done   int64
	notify func(int64)
	last   time.Time
}

func (c *countedReader) Read(b []byte) (int, error) {
	n, e := c.r.Read(b)
	c.done += int64(n)
	if time.Since(c.last) > 250*time.Millisecond {
		c.notify(c.done)
		c.last = time.Now()
	}
	return n, e
}
func (c *Client) Upload(ctx context.Context, path, remote string, notify func(int64)) error {
	if len(remote) > 100 {
		return errors.New("Invalid job filename")
	}
	for _, ch := range remote {
		if !(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '-' || ch == '.') {
			return errors.New("Invalid job filename")
		}
	}
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil || st.Size() < 1000 {
		return errors.New("The sliced archive is empty or unavailable")
	}
	connection, closeAll, e := c.ftp(ctx)
	if e != nil {
		return e
	}
	defer closeAll()
	defer connection.Quit()
	temporary := remote + ".part"
	reader := &countedReader{r: f, notify: notify}
	if e = connection.Stor(temporary, reader); e != nil {
		return errors.New("Transfer stopped before completion. The printer was not asked to start; check Wi-Fi, SD storage and the A1 data-channel mode")
	}
	if reader.done != st.Size() {
		return errors.New("Transfer byte count did not match the sliced archive; print start blocked")
	}
	size, e := connection.FileSize(temporary)
	if e != nil || size != st.Size() {
		return errors.New("Printer did not confirm the complete uploaded file size; print start blocked")
	}
	if e = connection.Rename(temporary, remote); e != nil {
		return errors.New("Printer storage could not finalize this job; print start blocked")
	}
	notify(st.Size())
	return nil
}
