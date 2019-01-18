package main

import (
	"github.com/hbahadorzadeh/wit/model"
	"github.com/hbahadorzadeh/wit/service"
	"github.com/janeczku/go-ipset/ipset"
	"go.uber.org/dig"
	"os"
)

func BuildContainer(args []string) *dig.Container {
	container := dig.New()

	//Config
	container.Provide(func() model.Config {
		return model.BuildConfigs(args)
	})

	//IpsetService
	ipsetServiceInstance := service.IpsetService{}
	container.Provide(func(config model.Config) *ipset.IPSet {
		return ipsetServiceInstance.GetInstance(config)
	})

	//WebService
	container.Provide(func(config model.Config, ipset *ipset.IPSet) *service.WebService {
		return service.GetWebService(config, ipset)
	})

	return container
}

func main() {
	container := BuildContainer(os.Args[1:])
	container.Invoke(func(webService service.WebService) {
		webService.Start()
	})
}
