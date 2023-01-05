package spdy

import (
	"bytes"
	"context"
	"io"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type stream struct {
	id     uint32
	mux    *muxer
	syn    atomic.Bool // 是否握手
	wmu    sync.Locker // 写锁
	cond   *sync.Cond
	buf    *bytes.Buffer
	err    error       // 错误信息
	closed atomic.Bool // 保证 close 方法只被执行一次
	ctx    context.Context
	cancel context.CancelFunc
}

func (stm *stream) ReadFrom(r io.Reader) (int64, error) {
	stm.cond.L.Lock()
	defer stm.cond.L.Unlock()

	if err := stm.err; err != nil {
		return 0, err
	}

	n, err := stm.buf.ReadFrom(r)

	stm.cond.Broadcast()

	return n, err
}

func (stm *stream) Read(p []byte) (n int, err error) {
	stm.cond.L.Lock()
	for {
		if buf := stm.buf; buf.Len() != 0 {
			n, err = buf.Read(p)
			break
		}
		if err = stm.err; err != nil {
			break
		}
		stm.cond.Wait()
	}
	stm.cond.L.Unlock()

	return
}

func (stm *stream) Write(b []byte) (int, error) {
	const max = math.MaxUint16
	bsz := len(b)
	if bsz == 0 {
		return 0, nil
	}

	flag := flagSYN
	if !stm.syn.CompareAndSwap(false, true) {
		flag = flagDAT
	}

	stm.wmu.Lock()
	defer stm.wmu.Unlock()

	for bsz > 0 {
		n := bsz
		if n > max {
			n = max
		}

		if _, err := stm.mux.write(flag, stm.id, b[:n]); err != nil {
			return 0, err
		}

		flag = flagDAT
		b = b[n:]
		bsz = len(b)
	}

	return bsz, nil
}

func (stm *stream) ID() uint32                       { return stm.id }
func (stm *stream) LocalAddr() net.Addr              { return stm.mux.LocalAddr() }
func (stm *stream) RemoteAddr() net.Addr             { return stm.mux.RemoteAddr() }
func (stm *stream) SetDeadline(time.Time) error      { return nil }
func (stm *stream) SetReadDeadline(time.Time) error  { return nil }
func (stm *stream) SetWriteDeadline(time.Time) error { return nil }

func (stm *stream) Close() error {
	return stm.closeError(io.EOF)
}

func (stm *stream) receive(p []byte) (int, error) {
	stm.cond.L.Lock()
	n, err := stm.buf.Write(p)
	stm.cond.L.Unlock()

	stm.cond.Broadcast()

	return n, err
}

func (stm *stream) closeError(err error) error {
	if !stm.closed.CompareAndSwap(false, true) {
		return io.ErrClosedPipe
	}

	stm.cond.L.Lock()
	stm.err = err
	stm.cond.L.Unlock()

	stm.cancel()

	stm.cond.Broadcast()

	return err
}
