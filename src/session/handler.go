package session

import (
	"encoding/binary"
	"io"
	"net"
)

func handlePacket(s *Session, buf []byte) error {
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

func handleHandShake(s *Session, buf []byte) error {
	_, err := io.ReadFull(s.dst.in, buf[:12])
	if err != nil {
		return err
	}

	targetAddr := bytesToAddr(buf[6:])
	conn, err := net.Dial("tcp4", targetAddr.String())
	if err != nil {
		// TODO
		sta := StateConnectionRefused
		h := genHeader(s, TypeHandShakeReply, sta)
		write(h, s.src.out)
		return err
	}
	defer conn.Close()
	s.src = &stream{
		in:   conn,
		out:  conn,
		addr: targetAddr,
	}
	h := genHeader(s, TypeHandShakeReply, StateSuccess)
	err = write(h, s.src.out)
	go s.listenTCP()
	return nil
}

func handleAlive(s *Session, buf []byte) error {
	h := genHeader(s, TypeAliveReply, StateSuccess)
	return write(h, s.src.out)
}

//TODO
func handleServiceHandShake(s *Session, buf []byte) error {
	return nil
}
