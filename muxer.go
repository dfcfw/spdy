package spdy

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
)

type Muxer interface {
	net.Listener

	LocalAddr() net.Addr

	RemoteAddr() net.Addr

	Open() (net.Conn, error)
}

type muxer struct {
	tran    net.Conn
	stmID   atomic.Uint32
	mutex   sync.RWMutex
	streams map[uint32]Streamer
	accepts chan Streamer
	closed  atomic.Bool
	done    chan struct{}
}

func (mux *muxer) LocalAddr() net.Addr {
	return mux.tran.LocalAddr()
}

func (mux *muxer) RemoteAddr() net.Addr {
	return mux.tran.RemoteAddr()
}

func (mux *muxer) Open() (net.Conn, error) {
	//TODO implement me
	panic("implement me")
}

func (mux *muxer) Accept() (net.Conn, error) {
	select {
	case stm := <-mux.accepts:
		return stm, nil
	case <-mux.done:
		return nil, io.ErrClosedPipe
	}
}

func (mux *muxer) Addr() net.Addr {
	return mux.LocalAddr()
}

func (mux *muxer) Close() error {
	if mux.closed.CompareAndSwap(false, true) {
		close(mux.done)
		return mux.tran.Close()
	}

	return io.ErrClosedPipe
}
