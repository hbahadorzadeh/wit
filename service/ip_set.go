package service

import (
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/hbahadorzadeh/wit/model"
	"github.com/janeczku/go-ipset/ipset"
	"log"
	"strconv"
	"sync"
)

type IpsetService struct {
	once  sync.Once
	value *ipset.IPSet
}

func (s *IpsetService) GetInstance(config model.Config) *ipset.IPSet {
	s.once.Do(func() {
		s.value = s.getIpsetService(config)
	})
	return s.value
}

func (is *IpsetService) getIpsetService(config model.Config) *ipset.IPSet {
	res, err := ipset.New(config.ListName, "hash:ip", &ipset.Params{})
	//iptables -A INPUT -m set ! --match-set WhiteList src  -d ip -p tcp --dport 80 -j REDIRECT --to-port %d
	ipt, err := iptables.New()
	iptList, err := ipt.List("nat", "PREROUTING")
	for _, port := range config.CoveringPorts {
		if err != nil {
		}
		flag := false
		dstPort := config.HttpsPort
		if port == 80 {
			dstPort = config.HttpPort
		}
		for _, rule := range iptList {
			if rule == fmt.Sprintf("-A PREROUTING -d %s/32 -p tcp -m set ! --match-set WhiteList src -m tcp --dport %d -j REDIRECT --to-ports %d", config.Bind, port, dstPort) {
				flag = true
				log.Printf("Rule for port %d was present\n", port)
				break
			}
		}
		if !flag {
			err := ipt.Append("nat", "PREROUTING", "-m", "set", "!", "--match-set", config.ListName, "src", "-d", config.Bind, "-p", "tcp", "--dport", strconv.FormatInt(int64(port), 10), "-j", "REDIRECT", "--to-port", strconv.FormatInt(int64(dstPort), 10))
			if err != nil {
				log.Printf("Failed to add rule for port %d\n%v\n", port, err)
			} else {
				log.Printf("Rule add for port %d\n", port)
			}
		}
	}
	if err != nil {
		log.Fatal(err)
	}
	return res
}
