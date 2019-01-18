package service

import (
	"crypto/tls"
	"fmt"
	"github.com/hbahadorzadeh/wit/model"
	"github.com/janeczku/go-ipset/ipset"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	htmlIndex = `<html><body>Welcome!</body></html>`
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, htmlIndex)
}

func handleLogin(ipset *ipset.IPSet) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["psk"]

		if !ok || len(keys[0]) < 1 {
			io.WriteString(w, "Url Param 'psk' is missing")
			return
		}

		// Query()["key"] will return an array of items,
		// we only want the single item.
		key := keys[0]

		log.Printf("key `%s` from ip `%s`: ", string(key), r.RemoteAddr)
		ipset.AddOption(r.RemoteAddr, "", 6*60*60)
		io.WriteString(w, "Wellcome!")
	}
}

type WebService struct {
	server      *http.Server
	certManager autocert.Manager
	config      model.Config
}

func GetWebService(config model.Config, ipset *ipset.IPSet) *WebService {
	server := WebService{}
	server.config = config

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/login/", handleLogin(ipset))

	os.MkdirAll(config.CertDir, 0700)

	server.certManager = autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(config.Host),
		Cache:      autocert.DirCache(config.CertDir),
	}

	server.server = &http.Server{
		Addr: fmt.Sprintf("%s:%d", config.Bind, config.HttpsPort), // e.g. you may want to listen on a high port
		TLSConfig: &tls.Config{
			GetCertificate: server.certManager.GetCertificate,
		},
		Handler: mux,
	}

	return &server
}

func (wb *WebService) Start() {

	go func() {
		h := wb.certManager.HTTPHandler(nil)
		log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", wb.config.Bind, wb.config.HttpPort), h))
	}()
	log.Fatal(wb.server.ListenAndServeTLS("", ""))
}
