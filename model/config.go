package model

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func help(err error) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(`
Useage:
	wit [optiosn]
	options:
		-a(auto_cert)
        -b bind_address
		-s server_address
		-l ListName
		-c CertDir
		-p http_port 
		-tp https_port
		-cp CoveringPorts(Comma seprated)
		
`)
	os.Exit(1)
}

type Config struct {
	AutoCert      bool
	Host          string
	Bind          string
	CertDir       string
	ListName      string
	HttpsKey      string
	HttpsCert     string
	HttpPort      int
	HttpsPort     int
	CoveringPorts []int
}

func BuildConfigs(args []string) Config {
	autoCert := false
	host := ""
	bind := "127.0.0.1"
	certDir := ""
	httpsKey := ""
	httpsCert := ""
	listName := "WhiteList"
	httpPort := 8001
	httpsPort := 8002
	coveringPorts := []int{80, 443, 1194, 8388}
	for i, arg := range args {
		switch arg {
		case "-a":
			autoCert = true
			break
		case "-s":
			host = args[i+1]
			break
		case "-b":
			bind = args[i+1]
			break
		case "-l":
			listName = args[i+1]
			break
		case "-c":
			certDir = args[i+1]
			break
		case "-p":
			portStr := args[i+1]
			port, err := strconv.Atoi(portStr)
			if err == nil {
				httpPort = port
			} else {
				help(errors.New(fmt.Sprintf("Invalid HttpPort(%s)", portStr)))
			}
			break
		case "-tp":
			portStr := args[i+1]
			port, err := strconv.Atoi(portStr)
			if err == nil {
				httpsPort = port
			} else {
				help(errors.New(fmt.Sprintf("Invalid HttpsPort(%s)", portStr)))
			}
			break
		case "-cp":
			portsStr := args[i+1]
			coveringPorts = []int{}
			for _, portStr := range strings.Split(portsStr, ",") {
				port, err := strconv.Atoi(portStr)
				if err == nil {
					coveringPorts = append(coveringPorts, port)
				} else {
					help(errors.New(fmt.Sprintf("Invalid HttpPort(%s)", portStr)))
				}
			}
			break
		default:
			break
		}
	}
	if autoCert == true {
		if host == "" {
			help(errors.New("Hostname most be specified for auto cert!"))
		}
		if bind == "127.0.0.1" {
			help(errors.New("Bind ip address most be valid internet ip for auto cert!"))
		}
		if certDir == "" {
			certDir = "cert"
		}
	} else {
		if certDir == "" {
			log.Println("No cert dir specified! making `cert` dir and key and cert file!")
			certDir = "cert"
		}

		files, err := ioutil.ReadDir(certDir)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if file.Name() == "server.key" {
				httpsKey = "server.key"
			}
			if file.Name() == "server.cert" {
				httpsCert = "server.cert"
			}
		}
		if httpsKey == "" || httpsCert == "" {
			help(errors.New("`server.key` or `server.cert` is not present in cert dir!"))
		}
	}
	return Config{
		AutoCert:      autoCert,
		Host:          host,
		Bind:          bind,
		CertDir:       certDir,
		ListName:      listName,
		HttpsCert:     httpsCert,
		HttpsKey:      httpsKey,
		HttpPort:      httpPort,
		HttpsPort:     httpsPort,
		CoveringPorts: coveringPorts,
	}
}
