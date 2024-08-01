package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func getRequest(r *http.Request, targetHostPort string, proto string) *http.Request {
	r.URL.Host = targetHostPort
	r.URL.Scheme = proto
	return &http.Request{
		Method: r.Method,
		URL: r.URL, //&url.URL{Scheme: proto, Host: targetHostPort, Path: r.URL},
		Header: r.Header,
		Body: r.Body,
		Proto: r.Proto,
		Form: r.Form,
		PostForm: r.PostForm,
		MultipartForm: r.MultipartForm,
		Trailer: r.Trailer,
	}
}


func HandleHTTP(w http.ResponseWriter, r *http.Request, targetHostPort string, proto string, writeToLog func(string) error) {  
	req := getRequest(r, targetHostPort, proto)
	
	writeToLog(fmt.Sprint("Request from Remoute: ", req))
	log.Println("Request from Remoute: ", req)

	resp, err := http.DefaultTransport.RoundTrip(req)
	writeToLog(fmt.Sprint("Response from Target", resp, err))
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