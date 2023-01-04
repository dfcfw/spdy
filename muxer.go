package spdy

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

type muxer struct {
	tran    net.Conn
	stmID   atomic.Uint32
	mutex   sync.RWMutex
	streams map[uint32]Streamer
	accepts chan Streamer
	ctx     context.Context
	cancel  context.CancelFunc
}

func (mux *muxer) Addr() net.Addr       { return mux.LocalAddr() }
func (mux *muxer) LocalAddr() net.Addr  { return mux.tran.LocalAddr() }
func (mux *muxer) RemoteAddr() net.Addr { return mux.tran.RemoteAddr() }

func (mux *muxer) Dial() (net.Conn, error) {
	stm, err := mux.newStream()
	if err != nil {
		return nil, err
	}

	mux.putStream(stm)

	return stm, nil
}

func (mux *muxer) Accept() (net.Conn, error) {
	select {
	case stm := <-mux.accepts:
		return stm, nil
	case <-mux.ctx.Done():
		return nil, io.ErrClosedPipe
	}
}

func (mux *muxer) Close() error {
	mux.cancel()
	return mux.tran.Close()
}

// Range 循环
func (mux *muxer) Range(fn func(Streamer) bool) {
	// copy on read
	mux.mutex.RLock()
	streams := make(map[uint32]Streamer, len(mux.streams))
	for k, v := range mux.streams {
		streams[k] = v
	}
	mux.mutex.RUnlock()

	for _, stm := range streams {
		if !fn(stm) {
			break
		}
	}
}

func (mux *muxer) newStream() (Streamer, error) {
	select {
	case <-mux.ctx.Done():
		return nil, io.ErrClosedPipe
	default:
	}

	stmID := mux.stmID.Add(2)
	cond := sync.NewCond(new(sync.Mutex))
	ctx, cancel := context.WithCancel(mux.ctx)

	stm := &stream{
		id:     stmID,
		mux:    mux,
		wmu:    new(sync.Mutex),
		cond:   cond,
		buf:    new(bytes.Buffer),
		ctx:    ctx,
		cancel: cancel,
	}

	return stm, nil
}

func (mux *muxer) synStream(stmID uint32) Streamer {
	cond := sync.NewCond(new(sync.Mutex))
	ctx, cancel := context.WithCancel(mux.ctx)

	return &stream{
		id:     stmID,
		mux:    mux,
		wmu:    new(sync.Mutex),
		cond:   cond,
		buf:    new(bytes.Buffer),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (mux *muxer) putStream(stm Streamer) {
	id := stm.ID()
	mux.mutex.Lock()
	mux.streams[id] = stm
	mux.mutex.Unlock()
}

func (mux *muxer) getStream(id uint32) Streamer {
	mux.mutex.RLock()
	stm := mux.streams[id]
	mux.mutex.RUnlock()

	return stm
}

func (mux *muxer) read() {
	defer func() {
		recover()
		_ = mux.Close()
		close(mux.accepts)
	}()

	var header frameHeader
	for {
		_, err := io.ReadFull(mux.tran, header[:])
		if err != nil {
			break
		}

		stmID := header.streamID()
		size := header.size()
		flag := header.flag()
		if flag&flagSYN == flagSYN {
			stm := mux.synStream(stmID)
			mux.putStream(stm)
			mux.accepts <- stm
		}
		if flag&flagPSH == flagPSH {

		}
		if flag&flagFIN == flagFIN {

		}

		switch flag {
		case flagSYN:

			// 读取数据
			stm.Read()

		case flagFIN:
		case flagPSH:
		}
	}
}
