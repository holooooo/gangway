package session

import (
	"encoding/binary"
	"io"
	"net"
)

type handler func(s *Session, sta state, buf []byte) error

var handlerMap map[pType]handler

func init() {
	handlerMap = map[pType]handler{
		TypePacket:           handlePacket,
		TypeAlive:            handleAlive,
		TypeShutdown:         handleShutdown,
		TypeServiceHandShake: handleServiceHandShake,

		TypeHandShake:      handleHandShake,
		TypeHandShakeReply: handleHandShakeReply,
	}
}

func handleShutdown(s *Session, sta state, buf []byte) error {
	s.log.Info().Msgf("Session %v to %v shutdown by remote", s.dst, s.src)
	close(s.stop)
	return nil
}

func handlePacket(s *Session, sta state, buf []byte) error {
	// if recive packet but never handshake, break
	if s.src == nil {
		return NotHandShakeYetErr
	}

	_, err := io.ReadFull(s.dst.in, buf[:2])
	if err != nil {
		return err
	}

	pLen := int64(binary.BigEndian.Uint16(buf[:2]))
	_, err = io.CopyN(s.src.out, s.dst.in, pLen)
	if err != nil {
		return err
	}
	return nil
}

func handleHandShake(s *Session, sta state, buf []byte) error {
	_, err := io.ReadFull(s.dst.in, buf[:6])
	if err != nil {
		return err
	}

	targetAddr := bytesToAddr(buf[:6])
	s.log.Info().Msgf("recived handshake: target to %v", targetAddr)
	conn, err := net.Dial("tcp4", targetAddr.String())
	if err != nil {
		// TODO correct return different error
		s.log.Warn().Msgf("handshake failed: target to %v", targetAddr)
		return err
	}
	defer conn.Close()
	s.src = &stream{
		in:   conn,
		out:  conn,
		addr: targetAddr,
	}
	h := genHeader(s, TypeHandShakeReply, StateSuccess)
	return write(h, s.src.out)
}

func handleHandShakeReply(s *Session, sta state, buf []byte) error {
	s.handShakeCh <- sta
	return nil
}

func handleAlive(s *Session, sta state, buf []byte) error {
	h := genHeader(s, TypeAliveReply, StateSuccess)
	return write(h, s.src.out)
}

//TODO
func handleServiceHandShake(s *Session, sta state, buf []byte) error {
	return nil
}
