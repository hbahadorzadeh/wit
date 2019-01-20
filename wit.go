package main

import (
	"context"
	"github.com/hbahadorzadeh/wit/model"
	"github.com/hbahadorzadeh/wit/service"
	"github.com/janeczku/go-ipset/ipset"
	"go.uber.org/dig"
	"log"
	"os"
	"os/exec"
)

func BuildContainer(args []string) *dig.Container {
	container := dig.New()

	//Config
	container.Provide(func() model.Config {
		return model.BuildConfigs(args)
	})

	//IptablesService
	container.Provide(func(config model.Config) *service.IpTables {
		return service.GetIptablesService(config)
	})

	//IpsetService
	ipsetServiceInstance := service.IpsetService{}
	container.Provide(func(config model.Config, ipt service.IpTables) *ipset.IPSet {
		return ipsetServiceInstance.GetInstance(config, ipt)
	})

	//WebService
	container.Provide(func(config model.Config, ipset *ipset.IPSet) *service.WebService {
		return service.GetWebService(config, ipset)
	})

	return container
}

func main() {
	if os.Geteuid() != 0 {
		log.Println("You need root permission!")
		cmd := exec.CommandContext(context.Background(), "/usr/bin/sudo", os.Args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		os.Exit(0)
	} else {
		container := BuildContainer(os.Args[1:])
		err := container.Invoke(func(webService *service.WebService) {
			webService.Start()
		})

		if err != nil {
			panic(err)
		}
	}
	os.Exit(0)
}
