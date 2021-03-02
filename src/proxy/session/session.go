package session

import (
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
	HandShakeTimeOut = 2 * time.Second

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
	TypeDns              pType = 7
	TypeDnsReply         pType = 8
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
	if s.isStop() {
		return
	}

	h := genHeader(s, TypeShutdown, StateSuccess)
	err := write(h, s.dst.out)
	if err != nil {
		return
	}

	select {
	case <-time.After(HandShakeTimeOut):
	case <-s.shutdownCh:
	}
	if !s.isStop() {
		close(s.stop)
	}
}

func (s *Session) listenPorto() {
	var err error
	buf := make([]byte, MaxPacketLen+4)
	for {
		if s.isStop() {
			return
		}
		ptype, sta, re := parseHeader(s.dst.in, buf)
		if re != nil {
			if e, ok := (re).(UnsportVersionErr); ok {
				log.Warn().Err(e)
				continue
			}
			err = re
			break
		}

		if handler, ok := handlerMap[ptype]; ok {
			err = handler(s, sta, buf)
		} else {
			log.Info().Msgf("Session %v to %v received error type %v", s.dst, s.src, sta)
		}

		if err != nil {
			break
		}

	}
	if err != nil {
		if err != io.EOF {
			log.Debug().Err(err).Msgf("Session %v to %v", s.src, s.dst)
			s.throwError(err)
		}
		s.shutdown()
	}
}

func (s *Session) listenTCP() {
	var err error
	buf := make([]byte, MaxPacketLen+4)
	for {
		if s.isStop() {
			break
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
		if err != io.EOF {
			log.Warn().Err(err).Msgf("Session %v to %v", s.src, s.dst)
			s.throwError(err)
		}
		s.shutdown()
	}
}

func (s *Session) isStop() bool {
	select {
	case <-s.stop:
		return true
	default:
		return false
	}
}

func (s *Session) Listen() {
	go s.listenTCP()
	go s.listenPorto()
	err := s.HandShake()
	if err != nil {
		log.Err(err).Msg("handshake failed")
		s.shutdown()
	}
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
