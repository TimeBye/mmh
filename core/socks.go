package core

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/mritd/mmh/common"
)

// DefaultSocksListenAddr default listen address for dynamic port forwarding
const DefaultSocksListenAddr = "127.0.0.1:8080"

const (
	socksVersion5 = 0x05

	socksCmdConnect = 0x01

	socksAtypIPv4   = 0x01
	socksAtypDomain = 0x03
	socksAtypIPv6   = 0x04

	socksRepSucceeded        = 0x00
	socksRepHostUnreachable  = 0x04
	socksRepCmdNotSupported  = 0x07
	socksRepAtypNotSupported = 0x08
)

// DynamicTunnel starts a local socks5 proxy(like `ssh -D`), all connections are
// forwarded to the target address through the ssh tunnel of the given server(support proxy)
func DynamicTunnel(name, listenAddr string) {
	listenAddr = NormalizeSocksListenAddr(listenAddr)

	server, err := findServerByName(name)
	common.CheckAndExit(err)
	client, err := server.wrapperClient(false)
	common.CheckAndExit(err)

	listener, err := net.Listen("tcp", listenAddr)
	common.CheckAndExit(err)
	defer func() { _ = listener.Close() }()
	fmt.Printf("mmh socks5 proxy listen at %s, via [%s]\n", listenAddr, name)

	for {
		conn, err := listener.Accept()
		if !common.CheckErr(err) {
			continue
		}
		go serveSocks(conn, client.Dial)
	}
}

// NormalizeSocksListenAddr supports "8080", ":8080", "0.0.0.0:1080" formats,
// empty address returns DefaultSocksListenAddr
func NormalizeSocksListenAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return DefaultSocksListenAddr
	}
	if _, err := strconv.Atoi(addr); err == nil {
		return "127.0.0.1:" + addr
	}
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	return addr
}

// dialFunc abstracts target address dialing, it is ssh client.Dial in production
type dialFunc func(network, address string) (net.Conn, error)

// serveSocks handles a single socks5 client connection(only CONNECT is supported, same as `ssh -D`)
func serveSocks(clientConn net.Conn, dial dialFunc) {
	defer func() { _ = clientConn.Close() }()

	target, ok := socksHandshake(clientConn)
	if !ok {
		return
	}

	remoteConn, err := dial("tcp", target)
	if err != nil {
		common.PrintErr(err)
		_ = socksReply(clientConn, socksRepHostUnreachable)
		return
	}

	fmt.Printf("new connection %s => %s\n", clientConn.RemoteAddr(), target)
	if !common.CheckErr(socksReply(clientConn, socksRepSucceeded)) {
		_ = remoteConn.Close()
		return
	}
	connCopy(remoteConn, clientConn)
}

// socksHandshake negotiates with the client(no auth) and parses the CONNECT target address
func socksHandshake(conn net.Conn) (string, bool) {
	// greeting: VER NMETHODS METHODS...
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		common.PrintErr(err)
		return "", false
	}
	if header[0] != socksVersion5 {
		common.PrintErr(fmt.Errorf("unsupported socks version: %d", header[0]))
		return "", false
	}
	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(conn, methods); err != nil {
		common.PrintErr(err)
		return "", false
	}
	if _, err := conn.Write([]byte{socksVersion5, 0x00}); err != nil {
		common.PrintErr(err)
		return "", false
	}

	// request: VER CMD RSV ATYP DST.ADDR DST.PORT
	req := make([]byte, 4)
	if _, err := io.ReadFull(conn, req); err != nil {
		common.PrintErr(err)
		return "", false
	}
	if req[0] != socksVersion5 {
		common.PrintErr(fmt.Errorf("unsupported socks version: %d", req[0]))
		return "", false
	}
	if req[1] != socksCmdConnect {
		_ = socksReply(conn, socksRepCmdNotSupported)
		common.PrintErr(fmt.Errorf("unsupported socks command: %d", req[1]))
		return "", false
	}

	var host string
	switch req[3] {
	case socksAtypIPv4:
		addr := make([]byte, net.IPv4len)
		if _, err := io.ReadFull(conn, addr); err != nil {
			common.PrintErr(err)
			return "", false
		}
		host = net.IP(addr).String()
	case socksAtypIPv6:
		addr := make([]byte, net.IPv6len)
		if _, err := io.ReadFull(conn, addr); err != nil {
			common.PrintErr(err)
			return "", false
		}
		host = net.IP(addr).String()
	case socksAtypDomain:
		addrLen := make([]byte, 1)
		if _, err := io.ReadFull(conn, addrLen); err != nil {
			common.PrintErr(err)
			return "", false
		}
		addr := make([]byte, int(addrLen[0]))
		if _, err := io.ReadFull(conn, addr); err != nil {
			common.PrintErr(err)
			return "", false
		}
		host = string(addr)
	default:
		_ = socksReply(conn, socksRepAtypNotSupported)
		common.PrintErr(fmt.Errorf("unsupported socks address type: %d", req[3]))
		return "", false
	}

	portBs := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBs); err != nil {
		common.PrintErr(err)
		return "", false
	}
	return net.JoinHostPort(host, strconv.Itoa(int(binary.BigEndian.Uint16(portBs)))), true
}

// socksReply writes a socks5 reply, BND.ADDR/BND.PORT are zeroed(clients usually ignore them)
func socksReply(conn net.Conn, rep byte) error {
	_, err := conn.Write([]byte{socksVersion5, rep, 0x00, socksAtypIPv4, 0, 0, 0, 0, 0, 0})
	return err
}
