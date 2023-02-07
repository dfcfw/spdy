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
	syn    bool        // 是否已经发送了握手帧
	wmu    sync.Locker // 数据写锁
	cond   *sync.Cond
	buf    *bytes.Buffer // 消息缓冲池
	err    error         // 错误信息
	closed atomic.Bool   // 保证 close 方法只被执行一次
	ctx    context.Context
	cancel context.CancelFunc
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

	stm.wmu.Lock()
	defer stm.wmu.Unlock()

	flag := flagDAT
	if !stm.syn {
		stm.syn = true
		flag = flagSYN
	}

	n := bsz
	for n > 0 {
		if n > max {
			n = max
		}

		if _, err := stm.mux.write(flag, stm.id, b[:n]); err != nil {
			return 0, err
		}

		flag = flagDAT
		b = b[n:]
		n = len(b)
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
	return stm.closeError(io.EOF, true)
}

func (stm *stream) receive(p []byte) (int, error) {
	stm.cond.L.Lock()
	n, err := stm.buf.Write(p) // FIXME: 尚未实现流控
	stm.cond.L.Unlock()

	stm.cond.Broadcast()

	return n, err
}

func (stm *stream) closeError(err error, fin bool) error {
	if !stm.closed.CompareAndSwap(false, true) {
		return io.ErrClosedPipe
	}

	stmID := stm.id
	stm.mux.delStream(stmID)

	if fin && stm.syn {
		_, _ = stm.mux.write(flagFIN, stmID, nil)
	}

	stm.cond.L.Lock()
	stm.err = err
	stm.cond.L.Unlock()

	stm.cancel()
	stm.cond.Broadcast()

	return err
}
