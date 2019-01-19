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
	"strings"
)

const (
	htmlIndex = `<html><body>Welcome!</body></html>`
)

func handleIndex(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, htmlIndex)
}

func handleLogin(ipset *ipset.IPSet, config model.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["psk"]

		if !ok || len(keys[0]) < 1 {
			io.WriteString(w, "Url Param 'psk' is missing")
			return
		}

		// Query()["key"] will return an array of items,
		// we only want the single item.
		key := keys[0]
		if config.PresharedKey != "" && config.PresharedKey != key {
			io.WriteString(w, "PSK does not match!")
			return
		}
		log.Printf("key `%s` from ip `%s`: ", string(key), r.RemoteAddr)
		addr := strings.Split(r.RemoteAddr, ":")[0]
		err := ipset.Add(addr, 6*60*60)
		if err!= nil{
			log.Println(err)
			io.WriteString(w, "Failed!")
		}else {
			io.WriteString(w, "Wellcome!")
		}
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
	mux.HandleFunc("/login/", handleLogin(ipset, config))

	if config.AutoCert {
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
	} else {
		server.server = &http.Server{
			Addr:    fmt.Sprintf("%s:%d", config.Bind, config.HttpsPort), // e.g. you may want to listen on a high port
			Handler: mux,
		}
	}

	return &server
}

func (wb *WebService) Start() {
	if wb.config.AutoCert {
		go func() {
			h := wb.certManager.HTTPHandler(nil)
			log.Printf("Http redirecting server started on http://%s:%d\n", wb.config.Bind, wb.config.HttpPort)
			log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", wb.config.Bind, wb.config.HttpPort), h))
		}()
		log.Printf("Https server started on https://%s:%d\n", wb.config.Bind, wb.config.HttpsPort)
		log.Fatal(wb.server.ListenAndServeTLS("", ""))
	} else {
		log.Printf("Https server started on https://%s:%d\n", wb.config.Bind, wb.config.HttpsPort)
		log.Fatal(wb.server.ListenAndServeTLS(wb.config.CertDir+"/"+wb.config.HttpsCert, wb.config.CertDir+"/"+wb.config.HttpsKey))
	}
}
