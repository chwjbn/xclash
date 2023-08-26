package inbound

import (
	"net"

	C "github.com/chwjbn/xclash/constant"
	"github.com/chwjbn/xclash/context"
	"github.com/chwjbn/xclash/transport/socks5"
)

// NewSocket receive TCP inbound and return ConnContext
func NewSocket(target socks5.Addr, conn net.Conn, source C.Type) *context.ConnContext {
	metadata := parseSocksAddr(target)
	metadata.NetWork = C.TCP
	metadata.Type = source
	if ip, port, err := parseAddr(conn.RemoteAddr().String()); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = port
	}

	return context.NewConnContext(conn, metadata)
}
