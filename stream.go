package spdy

import (
	"bytes"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Streamer interface {
	net.Conn
	ID() uint32
}

type stream struct {
	id   uint32
	mux  Muxer
	syn  atomic.Bool // 是否握手
	wmu  sync.Mutex  // 写锁
	rmu  sync.Mutex
	cond sync.Cond
	buf  *bytes.Buffer
	err  error
}

func (stm *stream) read(p []byte) (int, error) {
	stm.rmu.Lock()
	defer stm.rmu.Unlock()

	for {
		if buf := stm.buf; buf.Len() != 0 {
			return buf.Read(p)
		}
		if err := stm.err; err != nil {
			return 0, err
		}
		stm.cond.Wait()
	}
}

func (stm *stream) write(p []byte) (int, error) {
	var syn bool
	if stm.syn.CompareAndSwap(false, true) {
		syn = true
	}

	psz := len(p)
	stm.wmu.Lock()
	defer stm.wmu.Unlock()

}

func (stm *stream) ID() uint32 {
	return stm.id
}

func (stm *stream) Read(b []byte) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (stm *stream) Write(b []byte) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (stm *stream) Close() error {
	//TODO implement me
	panic("implement me")
}

func (stm *stream) LocalAddr() net.Addr {
	return stm.mux.LocalAddr()
}

func (stm *stream) RemoteAddr() net.Addr {
	return stm.mux.RemoteAddr()
}

func (stm *stream) SetDeadline(t time.Time) error {
	return nil
}

func (stm *stream) SetReadDeadline(t time.Time) error {
	return nil
}

func (stm *stream) SetWriteDeadline(t time.Time) error {
	return nil
}
