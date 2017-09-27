package main

import (
	"github.com/itachizhu/api-gateway-go/proxy"
	"github.com/itachizhu/api-gateway-go/repository"
	"github.com/itachizhu/api-gateway-go/domain"
)

func main() {
	defer func() {
		repository.Close()
		domain.Close()
	}()
	proxy.New().Run(":9070")
}
