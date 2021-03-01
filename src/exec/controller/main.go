package main

import (
	"gangway/src/proxy"
	"gangway/src/settings"
	"io"
	"net"
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	settings.ContextType = settings.ContextTypeController
	log.Info().Msg("Starting Gangway Conterller")
	go listenLog()
	proxy.Serve(proxy.TypeController, "127.0.0.1", 9527)
}

func listenLog() {
	server, err := net.Listen("tcp4", "127.0.0.1:7529")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Warn().Err(err)
			continue
		}
		go func() {
			defer conn.Close()
			_, _ = io.Copy(os.Stdout, conn)
		}()
	}
}
