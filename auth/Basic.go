package auth

import (
	"net/http"
	"strings"
	"encoding/base64"
	"bytes"
	"io"
)

var unauthorizedMsg = []byte("407 Proxy Authentication Required")

func BasicUnauthorized(req *http.Request, realm string) *http.Response {
	return &http.Response{
		StatusCode: 407,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Request:    req,
		Header: http.Header{
			"Proxy-Authenticate": []string{"Basic realm=" + realm},
			"Proxy-Connection":   []string{"close"},
		},
		Body:          io.NopCloser(bytes.NewBuffer(unauthorizedMsg)),
		ContentLength: int64(len(unauthorizedMsg)),
	}
}

var proxyAuthorizationHeader = "Proxy-Authorization"

func Auth(req *http.Request, f func(user, passwd string) bool) bool {
	authheader := strings.SplitN(req.Header.Get(proxyAuthorizationHeader), " ", 2)
	req.Header.Del(proxyAuthorizationHeader)
	if len(authheader) != 2 || authheader[0] != "Basic" {
		return false
	}
	userpassraw, err := base64.StdEncoding.DecodeString(authheader[1])
	if err != nil {
		return false
	}
	userpass := strings.SplitN(string(userpassraw), ":", 2)
	if len(userpass) != 2 {
		return false
	}
	return f(userpass[0], userpass[1])
}