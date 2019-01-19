package model

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
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
		-psk PresharedKey		
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
	PresharedKey  string
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
	presharedKey := ""
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
		case "-psk":
			presharedKey = args[i+1]
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
			os.MkdirAll(certDir, 0700)
		}
	} else {
		if certDir == "" {
			log.Println("No cert dir specified! making `cert` dir and key and cert file!")
			certDir = "cert"
			files, err := ioutil.ReadDir(certDir)
			if err != nil {
				log.Print(err)
			} else {
				for _, file := range files {
					if file.Name() == "server.key" {
						httpsKey = "server.key"
					}
					if file.Name() == "server.cert" {
						httpsCert = "server.cert"
					}
				}
			}

			if httpsKey == "" || httpsCert == "" {
				os.MkdirAll(certDir, 0700)
				os.Remove(certDir + "/server.crt")
				os.Remove(certDir + "/server.key")
				log.Println("`server.key` or `server.cert` is not present in cert dir!")
				host := flag.String("host", "", "Comma-separated hostnames and IPs to generate a certificate for")
				validFor := flag.Duration("duration", 365*24*time.Hour, "Duration that certificate is valid for")
				isCA := flag.Bool("ca", false, "whether this cert should be its own Certificate Authority")
				rsaBits := flag.Int("rsa-bits", 2048, "Size of RSA key to generate. Ignored if --ecdsa-curve is set")
				priv, err := rsa.GenerateKey(rand.Reader, *rsaBits)
				if err != nil {
					help(err)
				}

				notBefore := time.Now()
				notAfter := notBefore.Add(*validFor)

				serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
				serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
				if err != nil {
					log.Fatalf("failed to generate serial number: %s", err)
				}

				template := x509.Certificate{
					SerialNumber: serialNumber,
					Subject: pkix.Name{
						Organization: []string{"Acme Co"},
					},
					NotBefore:             notBefore,
					NotAfter:              notAfter,
					KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
					ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
					BasicConstraintsValid: true,
				}
				hosts := strings.Split(*host, ",")
				for _, h := range hosts {
					if ip := net.ParseIP(h); ip != nil {
						template.IPAddresses = append(template.IPAddresses, ip)
					} else {
						template.DNSNames = append(template.DNSNames, h)
					}
				}

				if *isCA {
					template.IsCA = true
					template.KeyUsage |= x509.KeyUsageCertSign
				}

				derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
				if err != nil {
					log.Fatalf("Failed to create certificate: %s", err)
					os.Exit(1)
				}

				certOut, err := os.Create(certDir + "/server.crt")
				if err != nil {
					log.Fatalf("failed to open server.crt for writing: %s", err)
					os.Exit(1)
				}

				if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
					log.Fatalf("failed to write data to server.crt: %s", err)
					os.Exit(1)
				}

				if err := certOut.Close(); err != nil {
					log.Fatalf("error closing server.crt: %s", err)
					os.Exit(1)
				}
				httpsCert = "server.crt"
				log.Print("wrote server.crt\n")

				keyOut, err := os.OpenFile(certDir+"/server.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)

				if err != nil {
					log.Print("failed to open server.key for writing:", err)
					os.Exit(1)
				}

				if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
					log.Fatalf("failed to write data to server.key %s", err)
					os.Exit(1)
				}

				if err := keyOut.Close(); err != nil {
					log.Fatalf("error closing server.key: %s", err)
					os.Exit(1)
				}
				log.Print("wrote server.key\n")
				httpsKey = "server.key"
			}
		} else {
			files, err := ioutil.ReadDir(certDir)
			if err != nil {
				log.Fatal(err)
			} else {
				for _, file := range files {
					if file.Name() == "server.key" {
						httpsKey = "server.key"
					}
					if file.Name() == "server.cert" {
						httpsCert = "server.cert"
					}
				}
			}
			if httpsKey == "" || httpsCert == "" {
				help(errors.New("`server.key` or `server.cert` is not present in cert dir!"))
			}
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
		PresharedKey:  presharedKey,
	}
}
