package goproxy


import (
	"http"

)


type ProxyCtx struct {
	// Will contain the client request from the proxy
	// Req *http.Request
	// Will contain the remote server's response (if available. nil if the request wasn't send yet)
	// Resp         *http.Response
	// RoundTripper RoundTripper
	// will contain the recent error that occurred while trying to send receive or parse traffic
	Error error
	// A handle for the user to keep data in the context, from the call of ReqHandler to the
	// call of RespHandler
	UserData interface{}
	// Will connect a request to a response
	Session   int64
	// certStore CertStorage
	Proxy     *ProxyHttpServer
}