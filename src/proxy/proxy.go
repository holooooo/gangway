package proxy

import (
	"fmt"
	"gangway/src/proxy/pool"
	"gangway/src/session"
	"net"

	"github.com/rs/zerolog/log"
)

const (
	TypeClient     = "client"
	TypeController = "controller"
)

// Serve will Listen local tcp conn and forward content by pipe
func Serve(t, ip string, port int) {
	addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%v:%v", ip, port))
	if err != nil {
		panic(err)
	}
	server, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	switch t {
	case TypeClient:
		log.Info().Msgf("Gangway Client start.Listening port %v", port)
	case TypeController:
		log.Info().Msgf("Gangway Controller start.Listening port %v", port)
	}

	for {
		conn, err := server.Accept()
		log.Debug().Msgf("recived a new connection from %v to %v", conn.LocalAddr(), conn.RemoteAddr())
		if err != nil {
			log.Warn().Err(err).Msgf("Get Conn from %v:%v failed", ip, port)
		}
		go func(c net.Conn) {
			defer c.Close()
			switch t {
			case TypeClient:
				err = proxyClient(c)
			case TypeController:
				err = proxyController(c)
			}
			if err != nil {
				log.Warn().Err(err).Msgf("Proxy to %v failed", conn.RemoteAddr())
			}
		}(conn)
	}
}

func proxyClient(c net.Conn) error {
	p, err := pool.GetPipe()
	if err != nil {
		return err
	}
	defer pool.Release(p)

	s, err := session.NewClientSession(c, p)
	if err != nil {
		return err
	}
	s.Listen()
	return nil
}

func proxyController(c net.Conn) error {
	s, err := session.NewServerSession(c)
	if err != nil {
		return err
	}
	s.Serve()
	return nil
}
