package inbound

import (
	"net"

	C "github.com/chwjbn/xclash/constant"
	"github.com/chwjbn/xclash/context"
	"github.com/chwjbn/xclash/transport/socks5"
)

// NewHTTP receive normal http request and return HTTPContext
func NewHTTP(target socks5.Addr, source net.Addr, conn net.Conn) *context.ConnContext {
	metadata := parseSocksAddr(target)
	metadata.NetWork = C.TCP
	metadata.Type = C.HTTP
	if ip, port, err := parseAddr(source.String()); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = port
	}
	return context.NewConnContext(conn, metadata)
}
