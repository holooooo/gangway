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
	settings.ContextType = settings.ContextTypeClient
	service.Init()
	kube.Init()
	pool.Init()
	proxy.Serve(proxy.TypeClient, *settings.IP, *settings.Port)
}
