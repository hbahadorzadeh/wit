package model

import (
	"errors"
	"fmt"
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
	Host          string
	Bind          string
	CertDir       string
	ListName      string
	HttpPort      int
	HttpsPort     int
	CoveringPorts []int
}

func BuildConfigs(args []string) Config {
	host := ""
	bind := "127.0.0.1"
	certDir := "cert"
	listName := "WhiteList"
	httpPort := 8001
	httpsPort := 8002
	coveringPorts := []int{80, 443, 1194, 8388}
	for i, arg := range args {
		switch arg {
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
	return Config{
		Host:          host,
		Bind:          bind,
		CertDir:       certDir,
		ListName:      listName,
		HttpPort:      httpPort,
		HttpsPort:     httpsPort,
		CoveringPorts: coveringPorts,
	}
}
