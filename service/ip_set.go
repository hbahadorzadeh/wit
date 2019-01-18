package service

import (
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/hbahadorzadeh/wit/model"
	"github.com/janeczku/go-ipset/ipset"
	"github.com/prometheus/common/log"
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
	for _, port := range config.CoveringPorts {
		list, err := ipt.List("filter", "INPUT")
		if err != nil {
		}
		flag := false
		for _, rule := range list {
			if rule == fmt.Sprintf("-A INPUT -m set ! --match-set WhiteList src  -d %s -p tcp --dport %d -j REDIRECT --to-port %d", config.Bind, port, config.HttpsPort) {
				flag = true
				break
			}
		}
		if !flag {
			ipt.Append("filter", "INPUT", "-m", "set", "!", "--match-set", config.ListName, "src", "-d", config.Bind, "-p", "tcp", "--dport", string(port), "-j", "REDIRECT", "--to-port", string(config.HttpsPort))
		}
	}
	if err != nil {
		log.Error(err)
	}
	return res
}
