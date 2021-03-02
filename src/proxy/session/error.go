package session

import (
	"errors"
	"fmt"
	"time"
)

type HandShakeTimeOutErr struct {
	timeout time.Duration
}

func NewHandShakeTimeOutErr(timeout time.Duration) error {
	return HandShakeTimeOutErr{timeout}
}

func (err HandShakeTimeOutErr) Error() string {
	return fmt.Sprintf("handshake timeout, current value is %v", err.timeout)
}

type UnsportVersionErr struct {
	cur    uint8
	target uint8
}

func NewUnsportVersionErr(cur, target uint8) error {
	return &UnsportVersionErr{target, cur}
}

func (err UnsportVersionErr) Error() string {
	return fmt.Sprintf("Gangway proto version %v don't support version %v", err.cur, err.target)
}

var (
	NotHandShakeYetErr = errors.New("should handshake before recive packet")
	ErrorHandShakeType = errors.New("error handshake type")
)
