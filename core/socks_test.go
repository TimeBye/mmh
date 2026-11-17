package core

import (
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestNormalizeSocksListenAddr(t *testing.T) {
	cases := map[string]string{
		"":               DefaultSocksListenAddr,
		"8080":           "127.0.0.1:8080",
		":8080":          "127.0.0.1:8080",
		"0.0.0.0:1080":   "0.0.0.0:1080",
		"127.0.0.1:8080": "127.0.0.1:8080",
		" 8080 ":         "127.0.0.1:8080",
	}
	for in, want := range cases {
		if got := NormalizeSocksListenAddr(in); got != want {
			t.Errorf("NormalizeSocksListenAddr(%q) = %q, want %q", in, got, want)
		}
	}
}

func startEchoServer(t *testing.T) net.Addr {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() { _ = conn.Close() }()
				_, _ = io.Copy(conn, conn)
			}()
		}
	}()
	return ln.Addr()
}

func startSocksServer(t *testing.T) net.Conn {
	t.Helper()
	clientConn, serverConn := net.Pipe()
	t.Cleanup(func() { _ = clientConn.Close() })
	go serveSocks(serverConn, net.Dial)
	return clientConn
}

func socksGreeting(t *testing.T, conn net.Conn) {
	t.Helper()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := conn.Write([]byte{socksVersion5, 0x01, 0x00}); err != nil {
		t.Fatal(err)
	}
	resp := make([]byte, 2)
	if _, err := io.ReadFull(conn, resp); err != nil {
		t.Fatal(err)
	}
	if resp[0] != socksVersion5 || resp[1] != 0x00 {
		t.Fatalf("unexpected greeting response: %v", resp)
	}
}

func socksRequest(t *testing.T, conn net.Conn, req []byte) []byte {
	t.Helper()
	if _, err := conn.Write(req); err != nil {
		t.Fatal(err)
	}
	reply := make([]byte, 10)
	if _, err := io.ReadFull(conn, reply); err != nil {
		t.Fatal(err)
	}
	return reply
}

func TestServeSocksConnectIPv4(t *testing.T) {
	echoAddr := startEchoServer(t)
	conn := startSocksServer(t)
	socksGreeting(t, conn)

	host, portStr, err := net.SplitHostPort(echoAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(portStr)
	req := []byte{socksVersion5, socksCmdConnect, 0x00, socksAtypIPv4}
	req = append(req, net.ParseIP(host).To4()...)
	req = binary.BigEndian.AppendUint16(req, uint16(port))
	if reply := socksRequest(t, conn, req); reply[1] != socksRepSucceeded {
		t.Fatalf("unexpected reply: %v", reply)
	}

	msg := []byte("ping")
	if _, err := conn.Write(msg); err != nil {
		t.Fatal(err)
	}
	got := make([]byte, len(msg))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatal(err)
	}
	if string(got) != string(msg) {
		t.Fatalf("echo mismatch: got %q, want %q", got, msg)
	}
}

func TestServeSocksConnectDomain(t *testing.T) {
	echoAddr := startEchoServer(t)
	conn := startSocksServer(t)
	socksGreeting(t, conn)

	_, portStr, err := net.SplitHostPort(echoAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(portStr)
	req := []byte{socksVersion5, socksCmdConnect, 0x00, socksAtypDomain, byte(len("localhost"))}
	req = append(req, "localhost"...)
	req = binary.BigEndian.AppendUint16(req, uint16(port))
	if reply := socksRequest(t, conn, req); reply[1] != socksRepSucceeded {
		t.Fatalf("unexpected reply: %v", reply)
	}
}

func TestServeSocksCommandNotSupported(t *testing.T) {
	conn := startSocksServer(t)
	socksGreeting(t, conn)

	req := []byte{socksVersion5, 0x02, 0x00, socksAtypIPv4}
	if reply := socksRequest(t, conn, req); reply[1] != socksRepCmdNotSupported {
		t.Fatalf("unexpected reply: %v", reply)
	}
}

func TestServeSocksAtypNotSupported(t *testing.T) {
	conn := startSocksServer(t)
	socksGreeting(t, conn)

	req := []byte{socksVersion5, socksCmdConnect, 0x00, 0x09}
	if reply := socksRequest(t, conn, req); reply[1] != socksRepAtypNotSupported {
		t.Fatalf("unexpected reply: %v", reply)
	}
}
