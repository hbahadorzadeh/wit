# wit
"Who Is There" or "WIT" is a linux app using iptables and ipset to cover your precious services. In fact it's a "Port knocking" server. It waits for your knock and if it recognize you, it will let you in!

```
Useage:
	wit [optiosn]
	options:
		-h,--help
		-version
		-a,--auto-cert
		-b,--bind-address bind_address
		-H,--host-name host_name
		-l,--list-name ListName
		-P,--policy redirect or drop
		-c,--cert-path CertPath
		-p,--http-port http_port
		-tp,--tls-port https_port
		-cp,--covering-ports CoveringPorts
		-psk PresharedKey	
```
It creates an ipset list (default name :"WhiteList") and adds iptables rules for each given CoveringPorts.
So it redirects all traffic to wit! Then you can authenticate by a HTTP_GET request like below and boom! You can reach your service for 6 hours :)
```
https://YOUR_BIND_IP/login/?pks=YOUR_PRESHARED_KEY
```
If you do not set any psk user will be authenticated by any key.
