package main

import (
	"github.com/itachizhu/api-gateway-go/proxy"
	"github.com/itachizhu/api-gateway-go/repository"
)

func main() {
	defer func() {
		repository.Close()
	}()
	proxy.New().Run(":9070")
}
