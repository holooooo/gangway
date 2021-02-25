package session

import (
	"encoding/binary"
	"gangway/src/kube/service"
	"gangway/src/proxy/pool"
	"io"
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

// state and type both use 4 bit
type state uint8
type pType uint8

const (
	MaxPacketLen     = int(^uint16(0))
	currentVersion   = uint8(0x1)
	HandShakeTimeOut = 10 * time.Second

	StateSuccess               state = 0
	StateNetworkUnreachable    state = 1
	StateIPUnreachable         state = 2
	StateConnectionRefused     state = 3
	StateTTLTimeout            state = 4
	StateUnsportProto          state = 5
	StateGangwayUnsportVersion state = 6

	TypePacket           pType = 0
	TypeHandShake        pType = 1
	TypeServiceHandShake pType = 2
	TypeHandShakeReply   pType = 3
	TypeShutdown         pType = 4
	TypeAlive            pType = 5
	TypeAliveReply       pType = 6
)

type Session struct {
	src *stream
	dst *stream

	stop        chan struct{}
	handShakeCh chan state
	shutdownCh  chan state
}

type stream struct {
	addr *net.TCPAddr
	in   io.Reader
	out  io.Writer
}

func (s stream) String() string {
	return s.addr.String()
}

func NewClientSession(c net.Conn, p pool.Pipe) (*Session, error) {
	src, err := net.ResolveTCPAddr("tcp4", c.LocalAddr().String())
	if err != nil {
		return nil, err
	}
	dst, err := net.ResolveTCPAddr("tcp4", c.RemoteAddr().String())
	if err != nil {
		return nil, err
	}

	s := &Session{
		src: &stream{
			addr: src,
			in:   c,
			out:  c,
		},
		dst: &stream{
			addr: dst,
			in:   p,
			out:  p,
		},
		stop:        make(chan struct{}),
		handShakeCh: make(chan state),
		shutdownCh:  make(chan state),
	}
	return s, nil
}

// server session need to listen handshake, and manage proxy conn
func NewServerSession(c net.Conn) (*Session, error) {
	dst, err := net.ResolveTCPAddr("tcp4", c.LocalAddr().String())
	if err != nil {
		return nil, err
	}
	s := &Session{
		stop:        make(chan struct{}),
		handShakeCh: make(chan state),
		shutdownCh:  make(chan state),
		dst: &stream{
			addr: dst,
			in:   c,
			out:  c,
		},
	}

	return s, nil
}

// todo need a service struct
func NewServiceSession(p *pool.Pipe, svc service.Service) (*Session, error) {
	return nil, nil
}

func (s *Session) HandShake() error {
	h := genHeader(s, TypeHandShake, StateSuccess)
	err := write(h, s.dst.out)
	if err != nil {
		return err
	}

	select {
	case <-time.After(HandShakeTimeOut):
		return NewHandShakeTimeOutErr(HandShakeTimeOut)
	case ss := <-s.handShakeCh:
		if ss == StateSuccess {
			return nil
		}
		return s.handleError(ss)
	}
}

// TODO
func (s *Session) ServiceHandShake(svc service.Service) error {
	return nil
}

// it is only called in tcp conn broken
func (s *Session) shutdown() {
	select {
	case <-s.stop:
		return
	default:
	}

	h := genHeader(s, TypeShutdown, StateSuccess)
	err := write(h, s.dst.out)
	if err != nil {
		return
	}

	select {
	case <-time.After(HandShakeTimeOut):
		log.Error().Msgf("Shutdown timeout")
	case <-s.shutdownCh:
	}
	close(s.stop)
}

func (s *Session) listenPorto() {
	var err error
	buf := make([]byte, MaxPacketLen+4)
	for {
		select {
		case <-s.stop:
			return
		default:
		}
		ptype, sta, re := parseHeader(s.dst.in, buf)
		if e, ok := (re).(UnsportVersionErr); ok {
			log.Err(e)
			continue
		} else if re != nil {
			err = re
			break
		}

		switch ptype {
		case TypeHandShakeReply:
			s.handShakeCh <- sta
		case TypeShutdown:
			log.Info().Msgf("Session %v to %v shutdown by remote", s.dst, s.src)
			close(s.stop)
		case TypePacket:
			// if recive packet but never handshake, break
			if s.src == nil {
				err = NotHandShakeYetErr
				break
			}

			_, re = io.ReadFull(s.dst.in, buf[:2])
			if re != nil {
				err = re
				break
			}

			pLen := int64(binary.BigEndian.Uint16(buf[:2]))
			_, pe := io.CopyN(s.src.out, s.dst.in, pLen)
			if pe != nil {
				err = pe
			}
		case TypeHandShake:
			_, re = io.ReadFull(s.dst.in, buf[:12])
			if re != nil {
				err = re
				break
			}

			targetAddr := bytesToAddr(buf[6:])
			conn, ce := net.Dial("tcp4", targetAddr.String())
			if ce != nil {
				err = ce
				// TODO
				sta := StateConnectionRefused
				h := genHeader(s, TypeHandShakeReply, sta)
				write(h, s.src.out)
				break
			}
			defer conn.Close()
			s.src = &stream{
				in:   conn,
				out:  conn,
				addr: targetAddr,
			}
			h := genHeader(s, TypeHandShakeReply, StateSuccess)
			err = write(h, s.src.out)
		case TypeAlive:
			h := genHeader(s, TypeAliveReply, StateSuccess)
			err = write(h, s.src.out)
		case TypeServiceHandShake:
			//TODO
		default:
			log.Info().Msgf("Session %v to %v recived error type %v", s.dst, s.src, sta)
		}
		if err != nil {
			break
		}

	}
	if err != nil {
		log.Err(err).Msgf("Session %v to %v", s.dst, s.src)
		s.throwError(err)
		s.shutdown()
	}
}

func (s *Session) listenTCP() {
	var err error
	buf := make([]byte, MaxPacketLen+4)
	for {
		select {
		case <-s.stop:
			return
		default:
		}
		rl, re := s.src.in.Read(buf[:MaxPacketLen])
		if rl > 0 {
			pl := packet(buf[:rl], buf)
			ne := write(buf[:pl], s.dst.out)
			if ne != nil {
				err = ne
				break
			}
		}
		if re != nil {
			err = re
			break
		}
	}

	if err != nil {
		log.Err(err).Msgf("Session %v to %v", s.src, s.dst)
		s.throwError(err)
		s.shutdown()
	}
}

func (s *Session) Listen() {
	go s.listenTCP()
	go s.listenPorto()
	<-s.stop
}

func (s *Session) Serve() {
	go s.listenPorto()
	<-s.stop
}

// TODO
func (s *Session) handleError(st state) error {
	return nil
}

// TODO write error to stream
func (s *Session) throwError(e error) {
}
