package handlers

import (
	"net/http"
	"log"
	"net"
	"fmt"
	"io"
	"time"
)


func HandleTunneling(w http.ResponseWriter, r *http.Request, writeToLog func(string) error) {

	writeToLog(fmt.Sprint("Hijacking connection:", r.RemoteAddr, "->", r.URL.Host))
	log.Println("Hijacking connection:", r.RemoteAddr, "->", r.URL.Host)
	clientConn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		writeToLog(fmt.Sprint("Hijack error:", err))
		log.Println("Hijack error:", err)
	  http.Error(w, err.Error(), http.StatusInternalServerError)
	  return
	}
  
	writeToLog(fmt.Sprint("Connecting to:", r.URL.Host))
	log.Println("Connecting to:", r.URL.Host)
	targetConn, err := net.DialTimeout("tcp", r.URL.Host, 10*time.Second)
	if err != nil {
		writeToLog(fmt.Sprint("Connect error:", err))
		log.Println("Connect error:", err)
	  writeRawResponse(clientConn, http.StatusServiceUnavailable, r, writeToLog)
	  return
	}
  
	writeRawResponse(clientConn, http.StatusOK, r, writeToLog)
  
	writeToLog(fmt.Sprint("Transferring:", r.RemoteAddr, "->", r.URL.Host))
	log.Println("Transferring:", r.RemoteAddr, "->", r.URL.Host)
	go transfer(targetConn, clientConn)
	go transfer(clientConn, targetConn)
  }
  
  func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
  }
  
  func writeRawResponse(conn net.Conn, statusCode int, r *http.Request, writeToLog func(string) error) {
	if _, err := fmt.Fprintf(conn, "HTTP/%d.%d %03d %s\r\n\r\n", r.ProtoMajor,
	  r.ProtoMinor, statusCode, http.StatusText(statusCode)); err != nil {
		writeToLog(fmt.Sprint("Writing response failed:", err))
		log.Println("Writing response failed:", err)
	}
  }