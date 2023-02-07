package spdy

import (
	"context"
	"net"
)

type option struct {
	maxsize  int
	backlog  int
	capacity int
	server   bool
	passwd   []byte
}

type Option func(*option)

func WithBacklog(n int) Option {
	return func(opt *option) {
		opt.backlog = n
	}
}

func WithMaxsize(n int) Option {
	return func(opt *option) {
		opt.maxsize = n
	}
}

func WithCapacity(n int) Option {
	return func(opt *option) {
		opt.capacity = n
	}
}

func WithEncrypt(passwd []byte) Option {
	return func(opt *option) {
		opt.passwd = passwd
	}
}

func (opt option) muxer(tran net.Conn) *muxer {
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
	passwd := opt.passwd

	ctx, cancel := context.WithCancel(context.Background())
	mux := &muxer{
		tran:    tran,
		streams: make(map[uint32]*stream, capacity),
		accepts: make(chan *stream, backlog),
		passwd:  passwd,
		ctx:     ctx,
		cancel:  cancel,
	}
	if opt.server {
		mux.stmID.Add(1)
	}

	return mux
}
