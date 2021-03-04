package main

import (
	"io"
	"net"
	"os"

	"github.com/rs/zerolog"
)

func main() {
	logConn, err := net.Dial("tcp4", "localhost:7529")
	if err != nil {
		return
	}
	defer logConn.Close()
	log := zerolog.New(logConn).With().Int("pid", os.Getpid()).Logger()

	dataConn, err := net.Dial("tcp4", "localhost:9527")
	if err != nil {
		log.Warn().Err(err).Msg("Connect failed, repeater shutdown ...")
	}
	defer dataConn.Close()

	stop := make(chan struct{})
	forward := func(dst io.Writer, src io.Reader) {
		_, _ = io.Copy(dst, src)
		stop <- struct{}{}
	}
	go forward(dataConn, os.Stdin)
	go forward(os.Stdout, dataConn)

	log.Debug().Msg("repeater started")
	<-stop
}
