package http

import (
	"fmt"
	"github.com/chwjbn/xclash/component/auth"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/chwjbn/xclash/adapter/inbound"
	"github.com/chwjbn/xclash/common/cache"
	N "github.com/chwjbn/xclash/common/net"
	C "github.com/chwjbn/xclash/constant"
	authStore "github.com/chwjbn/xclash/listener/auth"
	"github.com/chwjbn/xclash/log"
)

func HandleConn(c net.Conn, in chan<- C.ConnContext, cache *cache.Cache) {

	//http客户端,http时有效
	var xHttpClient *http.Client=nil
	defer func() {
		if xHttpClient!=nil{
			xHttpClient.CloseIdleConnections()
		}
	}()

	conn := N.NewBufferedConn(c)

	keepAlive := true

	trusted:=false
	if cache==nil{
		trusted=true
	}

	for keepAlive {

		request, err := ReadRequest(conn.Reader())
		if err != nil {
			break
		}

		keepAlive = strings.TrimSpace(strings.ToLower(request.Header.Get("Proxy-Connection"))) == "keep-alive"

		request.RemoteAddr = conn.RemoteAddr().String()

		var authUser *auth.AuthUser
		var resp *http.Response

		if !trusted{
			resp,authUser = authenticate(request, cache)
			if resp==nil{
				trusted=true
			}
		}

		if trusted{

			//HTTPS连接
			if strings.EqualFold(request.Method,http.MethodConnect){

				// Manual writing to support CONNECT for http 1.0 (workaround for uplay client)
				if _, err = fmt.Fprintf(conn, "HTTP/%d.%d %03d %s\r\n\r\n", request.ProtoMajor, request.ProtoMinor, http.StatusOK, "Connection established"); err != nil {
					break // close connection
				}

				connCtx:=inbound.NewHTTPS(request, conn)
				connCtx.SetAuthUser(authUser)

				in <- connCtx

				return // hijack connection

			}


			//HTTP连接
			if xHttpClient==nil{
				xHttpClient=newHttpClient(c.RemoteAddr(), in,authUser)
			}

			host := request.Header.Get("Host")
			if host != "" {
				request.Host = host
			}

			request.RequestURI = ""
			removeHopByHopHeaders(request.Header)
			removeExtraHTTPHostPort(request)

			if request.URL.Scheme == "" || request.URL.Host == "" {
				resp = responseWith(request, http.StatusBadRequest)
			} else {
				resp, err = xHttpClient.Do(request)
				if err != nil {
					resp = responseWith(request, http.StatusBadGateway)
				}
			}

			removeHopByHopHeaders(resp.Header)
		}


		if keepAlive {
			resp.Header.Set("Proxy-Connection", "keep-alive")
			resp.Header.Set("Connection", "keep-alive")
			resp.Header.Set("Keep-Alive", "timeout=4")
		}

		resp.Close = !keepAlive

		err = resp.Write(conn)
		if err != nil {
			break // close connection
		}

	}

	conn.Close()
}

func authenticate(request *http.Request, cache *cache.Cache) (*http.Response,*auth.AuthUser) {

	var authUser *auth.AuthUser=nil

	authenticator := authStore.Authenticator()
	if authenticator != nil {
		credential := parseBasicProxyAuthorization(request)
		if credential == "" {
			resp := responseWith(request, http.StatusProxyAuthRequired)
			resp.Header.Set("Proxy-Authenticate", "Basic")
			return resp,authUser
		}

		var authed interface{}
		if authed = cache.Get(credential); authed == nil {
			user, pass, err := decodeBasicProxyAuthorization(credential)
			authed = err == nil && authenticator.Verify(user, pass)
			cache.Put(credential, authed, time.Minute)

			authUser=&auth.AuthUser{User: user,Pass: pass}
		}
		if !authed.(bool) {
			log.Infoln("Auth failed from %s", request.RemoteAddr)

			return responseWith(request, http.StatusForbidden),authUser
		}
	}

	return nil,authUser
}

func responseWith(request *http.Request, statusCode int) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Proto:      request.Proto,
		ProtoMajor: request.ProtoMajor,
		ProtoMinor: request.ProtoMinor,
		Header:     http.Header{},
	}
}
