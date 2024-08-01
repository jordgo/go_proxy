package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"tadbox.com/go-proxy/auth"
	"tadbox.com/go-proxy/handlers"
	"tadbox.com/go-proxy/rotatelogger"
)

var (
	logFile *rotatelogger.File
  targetHostPort string
  targetProto string
  userName string
  userPwd string
)

func onLogClose(path string, didRotate bool) {
	fmt.Printf("we just closed a file '%s', didRotate: %v\n", path, didRotate)
	if !didRotate {
		return
	}
	// process just closed file e.g. upload to backblaze storage for backup
	go func() {
		// if processing takes a long time, do it in background
	}()
}

func openLogFile(pathFormat string, onClose func(string, bool)) error {
	w, err := rotatelogger.NewFile(pathFormat, onLogClose)
	if err != nil {
		return err
	}
	logFile = w
	return nil
}

func closeLogFile() error {
	return logFile.Close()
}

func writeToLog(msg string) error {
  t := time.Now()
	_, err := logFile.Write([]byte(fmt.Sprint("\n", t.Format("2006-01-02 15:04:05"), " - ", msg)))
	return err
}

// type Configuration struct {
//   HostPort string `validate:"hostname_port"`
// }

func main() {
  logDir := "../../logs"

	pathFormat := filepath.Join(logDir, "2006-01-02.txt")
	err := openLogFile(pathFormat, onLogClose)
	if err != nil {
		log.Fatalf("openLogFile failed with '%s'\n", err)
	}
	defer closeLogFile()


  serverPort := os.Getenv("SERVER_PORT")
  serverProto := os.Getenv("SERVER_PROTO")
  pemPath := os.Getenv("SSL_PEM_PATH")
  keyPath := os.Getenv("SSL_KEY_PATH")

  targetHost := os.Getenv("TARGET_HOST")
  targetPort := os.Getenv("TARGET_PORT")
  targetProto = os.Getenv("TARGET_PROTO")

  userName = os.Getenv("USER_NAME")
  userPwd = os.Getenv("USER_PWD")

  if serverPort == "" {
    serverPort = "8080"
  }

  if serverProto == "" {
    serverProto = "http"
  }
  if serverProto != "http" && serverProto != "https" {
    writeToLog("Server Protocol must be either http or https")
    log.Fatal("Server Protocol must be either http or https")
  }

  if serverProto == "https" && (pemPath == "" || keyPath == "") {
    writeToLog("pem and key path must be specified")
    log.Fatal("pem and key path must be specified")
  }

  if targetProto == "" {
    targetPort = "https"
  }

  if targetProto != "http" && targetProto != "https" {
    writeToLog("Target Protocol must be either http or https")
    log.Fatal("Target Protocol must be either http or https")
  }


  if serverProto == "http" && targetHost == ""  {
    writeToLog("Only CONNECT method is supported or Specify host and port")
    log.Println("Only CONNECT method is supported or Specify host and port")
  }

  if targetPort != "" {
    targetPortValidated, err := strconv.Atoi(targetPort)
    if err != nil {
      writeToLog(fmt.Sprintf("Invalid Target Port, error: %v\n", err))
      fmt.Fprintf(os.Stderr, "Invalid Target Port, error: %v\n", err)
      os.Exit(1)
    }
    if targetPortValidated > 0 && targetPortValidated < 65535 {
      targetPort = fmt.Sprint(targetPortValidated)
    } else {
      writeToLog(fmt.Sprintf("Invalid Target Port: %s", targetPort))
      fmt.Fprintf(os.Stderr, "Invalid Target Port: %s", targetPort)
      os.Exit(1)
    }
  }

  if targetPort == "" {
    targetHostPort = targetHost
  } else {
    targetHostPort = fmt.Sprintf("%s:%s", targetHost, targetPort)
  }
  proxyListenAddress := fmt.Sprintf("0.0.0.0:%s", serverPort)

  writeToLog(fmt.Sprint("Target host:", targetHostPort))
  log.Println(fmt.Sprint("Target host:", targetHostPort))

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
  writeToLog(fmt.Sprint("Started proxy at:", proxyServer.Addr))
  log.Println("Started proxy at:", proxyServer.Addr)
  if serverProto == "http" {
    log.Fatal(proxyServer.ListenAndServe())
  } else {
    log.Fatal(proxyServer.ListenAndServeTLS(pemPath, keyPath))
  }
}

func connectHandler(w http.ResponseWriter, r *http.Request) {
  if ok := auth.Auth(r, func (usr, pwd string) bool {
    return usr == userName && pwd == userPwd
  }); !ok {
    writeToLog("Authorization failed")
    log.Print("Authorization failed")
    w.WriteHeader(http.StatusProxyAuthRequired)
    return
  }

  if r.Method == http.MethodConnect {
    handlers.HandleTunneling(w, r, writeToLog)
  } else {
    handlers.HandleHTTP(w, r, targetHostPort, targetProto, writeToLog)
  }
}


