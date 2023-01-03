package spdy

import "net"

func Server(tran net.Conn, opts ...Option) Muxer {
	opt := &option{server: true}
	for _, fn := range opts {
		fn(opt)
	}

	return newMux(tran, opt)
}

func Client(tran net.Conn, opts ...Option) Muxer {
	opt := new(option)
	for _, fn := range opts {
		fn(opt)
	}

	return newMux(tran, opt)
}

func newMux(tran net.Conn, opt *option) Muxer {
	backlog := opt.backlog
	maxsize := opt.maxsize
	capacity := opt.capacity
	if backlog < 0 {
		backlog = 0
	}
	if maxsize <= 0 {
		maxsize = 40960 // 40 KiB
	}
	if capacity <= 0 {
		capacity = 64
	}

	mux := &muxer{
		tran:    tran,
		streams: make(map[uint32]Streamer, capacity),
		accepts: make(chan Streamer, backlog),
		done:    make(chan struct{}),
	}
	if !opt.server {
		mux.stmID.Add(1)
	}

	return mux
}
