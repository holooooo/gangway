package session

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/rs/zerolog/log"
)

func addrToBytes(addr *net.TCPAddr) []byte {
	b := make([]byte, 6)
	copy(b[:4], addr.IP[12:])
	b[4] = uint8(addr.Port >> 8)
	b[5] = uint8(addr.Port)
	return b
}
func bytesToAddr(b []byte) *net.TCPAddr {
	binary.BigEndian.Uint16(b)
	addr := &net.TCPAddr{
		IP:   b[:4],
		Port: int(b[4])<<8 | int(b[5]),
	}
	return addr
}

func mergeTypeAndState(t pType, s state) byte {
	return byte(t<<4) | byte(s)
}

func pluckTypeAndState(data byte) (pType, state) {
	return pType(data) >> 4, (state(data) << 4) >> 4
}

func write(data []byte, writer io.Writer) error {
	l, err := writer.Write(data)
	log.Debug().Msgf("write content: %v", data)
	if err != nil {
		return err
	} else if l != len(data) {
		return io.ErrShortWrite
	}
	return nil
}

func packet(data, buffer []byte) int {
	l := len(data)

	copy(buffer[4:], buffer[0:])
	buffer[0] = 0x01
	buffer[1] = mergeTypeAndState(TypePacket, StateSuccess)
	binary.BigEndian.PutUint16(buffer[2:4], uint16(l))
	return l + 4
}

func genHeader(s *Session, t pType, state state) []byte {
	h := make([]byte, 2)
	h[0] = currentVersion
	h[1] = mergeTypeAndState(t, state)
	return h
}

func parseHeader(in io.Reader, buf []byte) (pType, state, error) {
	_, re := io.ReadFull(in, buf[:1])
	if re != nil {
		return 0, 0, re
	}
	if buf[0] != currentVersion {
		return 0, 0, NewUnsportVersionErr(currentVersion, buf[0])
	}
	_, re = io.ReadFull(in, buf[:1])
	if re != nil {
		return 0, 0, re
	}
	p, s := pluckTypeAndState(buf[0])
	return p, s, nil
}
