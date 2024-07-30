package goproxy

import (
	// "log"
	"net"
	"http"
)


// The basic proxy type. Implements http.Handler.
type ProxyHttpServer struct {
	// session variable must be aligned in i386
	// see http://golang.org/src/pkg/sync/atomic/doc.go#L41
	sess int64
	// KeepDestinationHeaders indicates the proxy should retain any headers present in the http.Response before proxying
	KeepDestinationHeaders bool
	// setting Verbose to true will log information on each request sent to the proxy
	// Verbose         bool
	// Logger          Logger
	NonproxyHandler http.Handler
	// reqHandlers     []ReqHandler
	// respHandlers    []RespHandler
	// httpsHandlers   []HttpsHandler
	Tr              *http.Transport
	// ConnectDial will be used to create TCP connections for CONNECT requests
	// if nil Tr.Dial will be used
	ConnectDial        func(network string, addr string) (net.Conn, error)
	ConnectDialWithReq func(req *http.Request, network string, addr string) (net.Conn, error)
	// CertStore          CertStorage
	// KeepHeader         bool
	// AllowHTTP2         bool
}