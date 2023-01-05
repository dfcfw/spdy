package spdy_test

import (
	"net"
	"testing"
	"time"

	"github.com/dfcfw/spdy"
)

func TestServer(t *testing.T) {
	lis, err := net.Listen("tcp", ":9999")
	if err != nil {
		t.Fatal(err)
	}

	for {
		cli, err := lis.Accept()
		if err != nil {
			t.Fatal(err)
		}

		mux := spdy.Server(cli)

		go serve(t, mux)
	}
}

func serve(t *testing.T, mux spdy.Muxer) {
	for {
		conn, err := mux.Accept()
		if err != nil {
			t.Error(err)
			break
		}

		go serveConn(t, conn)
	}
}

func serveConn(t *testing.T, conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 100)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			t.Error(err)
			break
		}

		t.Logf(">>>>> %s", buf[:n])
	}
}

func TestClient(t *testing.T) {
	cli, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		t.Fatal(err)
	}

	mux := spdy.Client(cli)

	go func() {
		conn, err := mux.Dial()
		if err != nil {
			t.Fatal(err)
		}

		conn.Write([]byte("Hello----"))
		conn.Write([]byte("Hello1111111111111-----------"))
		conn.Write([]byte("Hello2222222----------------"))

		time.Sleep(3 * time.Second)

		conn.Write([]byte("你好------------"))

		conn.Close()
	}()

	conn, err := mux.Dial()
	if err != nil {
		t.Fatal(err)
	}

	conn.Write([]byte("Hello"))
	conn.Write([]byte("Hello1111111111111"))
	conn.Write([]byte("Hello2222222"))

	time.Sleep(3 * time.Second)

	conn.Write([]byte("你好"))

	conn.Close()

	time.Sleep(10 * time.Second)

}
