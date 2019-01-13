package letmein

import (
	"crypto/tls"
	"errors"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	htmlIndex = `<html><body>Welcome!</body></html>`
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, htmlIndex)
}

func help(err error){
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(`
Useage:
	wit [optiosn]
	options:
		-s server_address
		-p http_port 
		-tp https_port
		
`)
	os.Exit(1)
}

func main() {

	//Defaults
	host := "my.server.me"
	cacheDir := "cert"
	httpPort := 8001
	httpsPort := 8002
	argsWithoutProg := os.Args[1:]

	for i, arg := range argsWithoutProg {
		switch arg {
		case "-s":
			host = argsWithoutProg[i+1]
			break
		case "-p":
			portStr := argsWithoutProg[i+1]
			port, err := strconv.Atoi(portStr)
			if(err== nil){
				httpPort = port
			}else{
				help(errors.New(fmt.Sprintf("Invalid httpPort(%s)", portStr)))
			}
			break
		case "-tp":
		case "":
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	os.MkdirAll(cacheDir, 0700)
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(host),
		Cache:      autocert.DirCache(cacheDir),
	}
	server := &http.Server{
		Addr: fmt.Sprintf("127.0.0.1:%d", httpsPort), // e.g. you may want to listen on a high port
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
		Handler: mux,
	}
	// add your listeners via http.Handle("/path", handlerObject)
	log.Fatal(server.ListenAndServeTLS("", ""))

	go func() {
		h := certManager.HTTPHandler(nil)
		log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", httpPort), h))
	}()
}
