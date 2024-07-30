package main

import (
	"crypto/tls"
	// "flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	// "log/slog"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

var targetHostPort string
var authToken string
// type Configuration struct {
//   HostPort string `validate:"hostname_port"`
// }

func main() {
  serverPort := os.Getenv("SERVER_PORT")
  proto := os.Getenv("PROTO")
  targetHost := os.Getenv("HOST")
  targetPort := os.Getenv("PORT")
  pemPath := os.Getenv("SSL_PEM_PATH")
  keyPath := os.Getenv("SSL_KEY_PATH")
  authToken = os.Getenv("TOKEN")

  if proto != "http" && proto != "https" {
    log.Fatal("Protocol must be either http or https")
  }

  if proto == "https" && (pemPath == "" || keyPath == "") {
    log.Fatal("pem and key path must be specified")
  }

  if proto == "http" && targetHost == "" || targetPort == "" {
    log.Println("Only CONNECT method is supported or Specify host and port")
  }

  if targetPort != "" {
    targetPortValidated, err := strconv.Atoi(targetPort)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Invalid Target Port, error: %v\n", err)
      os.Exit(1)
    }
    if targetPortValidated > 0 && targetPortValidated < 65535 {
      targetPort = fmt.Sprint(targetPortValidated)
    } else {
      fmt.Fprintf(os.Stderr, "Invalid Target Port: %s", targetPort)
      os.Exit(1)
    }
  }

  targetHostPort = fmt.Sprintf("%s:%s", targetHost, targetPort)
  proxyListenAddress := fmt.Sprintf("0.0.0.0:%s", serverPort)

  log.Println(targetHostPort)

  // c := Configuration{
  //   HostPort: fmt.Sprintf("%s:%s", targetHost, targetPort),
  // }
  // validator := validator.New()
  // var validate *validator.Validate
  // validate = validator.New()
  // validate.Struct(c)

  // assert.NoError(t, validator.Struct(c))

  proxyServer := http.Server{
    Addr:         proxyListenAddress,
    Handler:      http.HandlerFunc(connectHandler),
    TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){},
  }
  log.Println("Started proxy at:", proxyServer.Addr)
  if proto == "http" {
    log.Fatal(proxyServer.ListenAndServe())
  } else {
    log.Fatal(proxyServer.ListenAndServeTLS(pemPath, keyPath))
  }
}

func connectHandler(w http.ResponseWriter, r *http.Request) {
  if ok := checkToken(r); !ok {
    log.Print("Authorization failed")
    w.WriteHeader(http.StatusProxyAuthRequired)
    return
  }

  if r.Method == http.MethodConnect {
    handleTunneling(w, r)
  } else {
    handleHTTP(w, r)
  }
}

func checkToken(r *http.Request) bool {
  arr := strings.Split(r.Header.Get("Proxy-Authorization"), " ")

  //without auth
  if authToken == "" {
    log.Println("Without authorization")
    return true
  }

  //token missed
  if len(arr) == 0 {
    log.Panicln("Token not found")
    return false
  }

  return arr[len(arr) - 1] == authToken
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
  r.URL = &url.URL{Scheme: "http", Host: targetHostPort, Path: r.URL.String()}
  r.RequestURI = ""
  log.Println("Request from Remoute: ", r)

  resp, err := http.DefaultTransport.RoundTrip(r)
  log.Println("Response from Target", resp, err)

  if err != nil {
    http.Error(w, err.Error(), http.StatusServiceUnavailable)
    return
  }
  defer resp.Body.Close()

  copyHeader(w.Header(), resp.Header)

  w.WriteHeader(resp.StatusCode)
  io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
  for k, vv := range src {
    for _, v := range vv {
      dst.Add(k, v)
    }
  }
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {

  log.Println("Hijacking connection:", r.RemoteAddr, "->", r.URL.Host)
  clientConn, _, err := w.(http.Hijacker).Hijack()
  if err != nil {
    log.Println("Hijack error:", err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  log.Println("Connecting to:", r.URL.Host)
  targetConn, err := net.DialTimeout("tcp", r.URL.Host, 10*time.Second)
  if err != nil {
    log.Println("Connect error:", err)
    writeRawResponse(clientConn, http.StatusServiceUnavailable, r)
    return
  }

  writeRawResponse(clientConn, http.StatusOK, r)

  log.Println("Transferring:", r.RemoteAddr, "->", r.URL.Host)
  go transfer(targetConn, clientConn)
  go transfer(clientConn, targetConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
  defer destination.Close()
  defer source.Close()
  io.Copy(destination, source)
}

func writeRawResponse(conn net.Conn, statusCode int, r *http.Request) {
  if _, err := fmt.Fprintf(conn, "HTTP/%d.%d %03d %s\r\n\r\n", r.ProtoMajor,
    r.ProtoMinor, statusCode, http.StatusText(statusCode)); err != nil {
    log.Println("Writing response failed:", err)
  }
}