package main

import (
	"gangway/src/kube"
	"gangway/src/kube/service"
	"gangway/src/proxy"
	"gangway/src/proxy/pool"
	"gangway/src/settings"

	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Starting Gangway Client")
	svcLen, err := service.Collect()
	if err != nil {
		panic(err)
	}
	kube.StreamInit()
	pool.InitPool(*settings.PoolSize+svcLen, *settings.PoolMaxIdle)
	go service.Register()
	proxy.Serve(proxy.TypeClient, *settings.IP, *settings.Port)
}
