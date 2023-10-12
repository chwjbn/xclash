package http

import (
	"context"
	"errors"
	"github.com/chwjbn/xclash/component/auth"
	"net"
	"net/http"
	"time"

	"github.com/chwjbn/xclash/adapter/inbound"
	C "github.com/chwjbn/xclash/constant"
	"github.com/chwjbn/xclash/transport/socks5"
)

func newHttpClient(source net.Addr, in chan<- C.ConnContext,authUser *auth.AuthUser) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			// from http.DefaultTransport
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext: func(context context.Context, network, address string) (net.Conn, error) {
				if network != "tcp" && network != "tcp4" && network != "tcp6" {
					return nil, errors.New("unsupported network " + network)
				}

				dstAddr := socks5.ParseAddr(address)
				if dstAddr == nil {
					return nil, socks5.ErrAddressNotSupported
				}

				left, right := net.Pipe()

				xConnCtx:=inbound.NewHTTP(dstAddr, source, right)
				xConnCtx.SetAuthUser(authUser)

				in <- xConnCtx

				return left, nil
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
