package service

import (
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/hbahadorzadeh/wit/model"
	"log"
	"strconv"
	"strings"
	"sync"
)

type IPT_POLICY int

const (
	DROP       IPT_POLICY = 0
	REDIRECT   IPT_POLICY = 1
	CHAIN_NAME            = "WIT"
)

type IpTables struct {
	once sync.Once

	policy      IPT_POLICY
	table       string
	parentchain string

	listName  string
	bind      string
	httpsPort int

	ipt    *iptables.IPTables
	config model.Config
}

func GetIptablesService(config model.Config) *IpTables {
	return &IpTables{
		config: config,
	}
}

func (it *IpTables) Init() *IpTables {
	it.once.Do(func() {
		it.init(it.config)
	})
	return it
}

func (it *IpTables) init(config model.Config) {
	it.listName = config.ListName

	switch strings.ToLower(config.Policy) {
	case "drop":
		it.policy = DROP
		it.table = "filter"
		it.parentchain = "INPUT"
	case "redirect":
	default:
		it.policy = REDIRECT
		it.table = "nat"
		it.parentchain = "PREROUTING"
	}

	it.bind = config.Bind

	it.httpsPort = config.HttpsPort

	ipt, err := iptables.New()
	if err != nil {
		log.Panic(err)
	}
	it.ipt = ipt

	err = it.initChain()
	if err != nil {
		log.Panic(err)
	}

	for _, port := range config.CoveringPorts {
		err := ipt.Append(it.table, CHAIN_NAME, it.makeRule(port)...)
		if err != nil {
			log.Printf("Failed to add rule for port %d\n%v\n", port, err)
		} else {
			log.Printf("Rule add for port %d\n", port)
		}
	}
}

func (it *IpTables) initChain() error {

	_ = it.ipt.NewChain(it.table, CHAIN_NAME)
	_ = it.ipt.ClearChain(it.table, CHAIN_NAME)

	iptList, err := it.ipt.List(it.table, it.parentchain)
	if err != nil {
		return err
	}

	flag := false
	for _, iptRule := range iptList {
		if iptRule == fmt.Sprintf("-A %s -j %s", it.parentchain, CHAIN_NAME) {
			flag = true
			break
		}
	}

	if !flag {
		err = it.ipt.Append(it.table, it.parentchain, []string{"-j", CHAIN_NAME}...)
	} else {
		err = nil
	}
	return err
}

func (it IpTables) makeRule(dport int) []string {
	var rule []string

	if it.bind != "0.0.0.0" {
		rule = []string{"-m", "set", "!", "--match-set", it.listName, "src", "-d", it.bind, "-p", "tcp", "--dport", strconv.FormatInt(int64(dport), 10)}
	} else {
		rule = []string{"-m", "set", "!", "--match-set", it.listName, "src", "-p", "tcp", "--dport", strconv.FormatInt(int64(dport), 10)}
	}

	switch it.policy {
	case DROP:
		rule = append(rule, []string{"-j", "DROP"}...)
	case REDIRECT:
	default:
		rule = append(rule, []string{"-j", "REDIRECT", "--to-port", strconv.FormatInt(int64(it.httpsPort), 10)}...)
	}

	return rule
}

func (it *IpTables) Destroy(config model.Config) {
	_ = it.ipt.ClearChain(it.table, CHAIN_NAME)
	_ = it.ipt.DeleteChain(it.table, CHAIN_NAME)
	_ = it.ipt.Delete(it.table, it.parentchain, []string{"-j", CHAIN_NAME}...)
}
