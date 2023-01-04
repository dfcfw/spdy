package spdy

import (
	"testing"
)

func TestServer(t *testing.T) {
	fh := frameHeader{0: flagSYN | flagPSH | flagFIN | flagNOP, 1: 34}
	t.Log(fh.String())
}
