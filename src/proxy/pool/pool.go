package pool

import (
	"errors"
	"gangway/src/kube"
	"io"
	"sync"
	"time"
)

type Pipe interface {
	io.Reader
	io.Writer
	io.Closer
	Alive() error
}

var (
	m            sync.Mutex
	idlePipes    chan Pipe
	checkedPipes chan Pipe

	maxIdle  int
	size     int
	connNums int
)

var (
	ExceedMaxConn = errors.New("The connections number exceeds the set maximum")
)

const (
	IdleTimeOut time.Duration = 5 * time.Minute
)

func InitPool(s, mi int) {
	size = s
	maxIdle = mi
	idlePipes = make(chan Pipe, size)
	checkedPipes = make(chan Pipe, maxIdle)
	go func() {
		beat()
	}()
}

func GetPipe() (Pipe, error) {
	select {
	case conn := <-idlePipes:
		return conn, nil
	case <-time.After(2 * time.Second):
	}
	if connNums < size {
		m.Lock()
		defer m.Unlock()
		connNums++
		return kube.NewPipe()
	}
	return nil, ExceedMaxConn
}

func Release(p Pipe) {
	select {
	case idlePipes <- p:
		return
	default:
		m.Lock()
		defer m.Unlock()
		p.Close()
		connNums--
	}
}

func beat() {
	for {
		select {
		case <-time.After(10 * time.Second):
			check()
		}
	}
}

func check() {
	idleNums := 0
CheckIdle:
	for {
		select {
		case pipe := <-idlePipes:
			if idleNums == maxIdle || pipe.Alive() != nil {
				m.Lock()
				connNums--
				pipe.Close()
				m.Unlock()
				break
			}
			checkedPipes <- pipe
			idleNums++
		default:
			break CheckIdle
		}
	}
RealeaseChecked:
	for {
		select {
		case pipe := <-checkedPipes:
			Release(pipe)
		default:
			break RealeaseChecked
		}
	}
}
