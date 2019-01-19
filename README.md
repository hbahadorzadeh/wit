# wit
"Who Is There" or "WIT" is a linux app using iptables and ipset to cover your precious services. In fact it's a "Port knocking" server. It waits for your knock and if it recognize you, it will let you in!

```
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
```
It creates an ipset list (default name :"WhiteList") and adds iptables rules for each given CoveringPorts as below:
```
-t nat -A OUTPUT -d 127.0.0.1/32 -p tcp -m set ! --match-set WhiteList src 
            -m tcp --dport YOUR_SERVICE_PORT -j REDIRECT --to-ports WIT_HTTPS_PORT
```

So it redirects all traffic to wit! Then you can authenticate by a HTTP_GET request like below and boom! You can reach your service for 6 hours :)
```
https://YOUR_BIND_IP/login/?pks=YOUR_PRESHARED_KEY
```
If you do not set any psk user will be authenticate by any key.
