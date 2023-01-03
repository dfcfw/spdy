package spdy

type option struct {
	maxsize  int
	backlog  int
	capacity int
	server   bool
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
