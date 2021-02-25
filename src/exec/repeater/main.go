package main

import (
	"io"
	"net"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

var (
	log zerolog.Logger
	pid = os.Getpid()
)

func main() {
	logConn, err := net.Dial("tcp4", "localhost:9527")
	if err != nil {
		return
	}
	defer logConn.Close()

	log = zerolog.New(logConn).With().Str("pid", strconv.Itoa(pid)).Logger()

	dataConn, err := net.Dial("tcp4", "localhost:7529")
	if err != nil {
		log.Err(err).Msg("Connect failed, shutdown ...")
	}
	defer dataConn.Close()

	stop := make(chan struct{})
	forward := func(dst io.Writer, src io.Reader) {
		_, _ = io.Copy(dst, src)
		stop <- struct{}{}
	}
	go forward(dataConn, os.Stdin)
	go forward(os.Stdout, dataConn)
	<-stop
}
