package service

import (
	"github.com/hbahadorzadeh/wit/model"
	"github.com/janeczku/go-ipset/ipset"
	"log"
	"sync"
)

type IpsetService struct {
	once  sync.Once
	value *ipset.IPSet
}

func (s *IpsetService) GetInstance(config model.Config, ipts IpTables) *ipset.IPSet {
	s.once.Do(func() {
		s.value = s.getIpsetService(config, ipts)
	})
	return s.value
}

func (is *IpsetService) getIpsetService(config model.Config, ipts IpTables) *ipset.IPSet {
	res, err := ipset.New(config.ListName, "hash:ip", &ipset.Params{})
	if err != nil {
		log.Panic(err)
	}
	ipts.Init()
	return res
}
