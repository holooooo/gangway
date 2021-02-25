package main

import (
	"gangway/src/kube/service"
	"gangway/src/proxy"
	"gangway/src/proxy/pool"
	"gangway/src/settings"
)

func main() {
	svcLen, err := service.Collect()
	if err != nil {
		panic(err)
	}
	pool.InitPool(*settings.PoolSize+svcLen, *settings.PoolMaxIdle)
	go service.Register()
	proxy.Serve(proxy.TypeClient, *settings.IP, *settings.Port)
}
