package pool

import (
	"errors"
	"gangway/src/kube"
	"gangway/src/settings"
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

func Init() {
	size = *settings.PoolSize
	maxIdle = *settings.PoolMaxIdle
	idlePipes = make(chan Pipe, size)
	checkedPipes = make(chan Pipe, maxIdle)
	beat()
}

func GetPipe() (Pipe, error) {
	select {
	case conn := <-idlePipes:
		return conn, nil
	case <-time.After(2 * time.Second):
	}
	if connNums < size {
		return NewPipe()
	}
	return nil, ExceedMaxConn
}

func NewPipe() (Pipe, error) {
	m.Lock()
	defer m.Unlock()
	connNums++
	return kube.NewPipe()
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
	go func() {
		for {
			time.Sleep(10 * time.Second)
			check()
		}
	}()
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
	for {
		select {
		case pipe := <-checkedPipes:
			Release(pipe)
		default:
			return
		}
	}
}
