package spdy

import "net"

type Muxer interface {
	net.Listener

	LocalAddr() net.Addr

	RemoteAddr() net.Addr

	Dial() (Streamer, error)
}

type Streamer interface {
	net.Conn
	ID() uint32
}

func Server(tran net.Conn, opts ...Option) Muxer {
	opt := &option{server: true}
	for _, fn := range opts {
		fn(opt)
	}

	mux := opt.muxer(tran)
	go mux.read()

	return mux
}

func Client(tran net.Conn, opts ...Option) Muxer {
	opt := new(option)
	for _, fn := range opts {
		fn(opt)
	}

	mux := opt.muxer(tran)
	go mux.read()

	return mux
}
