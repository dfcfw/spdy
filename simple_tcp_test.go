package spdy_test

import (
	"net"
	"testing"
	"time"

	"github.com/dfcfw/spdy"
)

func Test_TCP_Server(t *testing.T) {
	lis, err := net.Listen("tcp", ":9999")
	if err != nil {
		t.Fatal(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer lis.Close()

	for {
		conn, ex := lis.Accept()
		if ex != nil {
			t.Log(err)
			break
		}

		go acceptConn(t, conn)
	}
}

func acceptConn(t *testing.T, conn net.Conn) {
	//goland:noinspection GoUnhandledErrorResult
	defer conn.Close()

	// 基于 TCP Socket 建立多路复用
	mux := spdy.Server(conn)

	go func() {
		// 服务端主动建立一个虚拟连接
		dial, err := mux.Dial()
		if err != nil {
			t.Log(err)
			return
		}
		//goland:noinspection GoUnhandledErrorResult
		defer dial.Close()

		// 启动一个线程定时发送消息
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case at := <-ticker.C:
				txt, _ := at.MarshalText()
				if _, ex := dial.Write(txt); ex == nil {
					t.Logf("[SRV virtual conn] 发送消息成功")
				} else {
					t.Logf("[SRV virtual conn] 发送消息错误：%v", ex)
					return
				}
			}
		}
	}()

	for {
		cli, err := mux.Accept() // 客户端建立虚拟连接
		if err != nil {
			t.Log(err)
			break
		}

		go serveConn(t, cli)
	}
}

func serveConn(t *testing.T, conn net.Conn) {
	defer func() {
		_ = conn.Close()
		t.Log("[SRV (virtual conn)] 已断开")
	}()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			t.Logf("[SRV (virtual conn)] 读取发生错误：%v", err)
			break
		}
		t.Logf("[SRV (virtual conn)] 收到消息: %s", buf[:n])
	}
}

func Test_TCP_Client(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		t.Fatal(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer conn.Close()

	// 基于 TCP Socket 建立多路复用
	mux := spdy.Client(conn)
	for {
		cli, ex := mux.Accept()
		if err != nil {
			t.Log(ex)
			break
		}

		go serveCli(t, cli)
	}

}

func serveCli(t *testing.T, conn net.Conn) {
	signals := make(chan struct{})
	defer func() {
		close(signals)
		_ = conn.Close()
		t.Log("[CLI (virtual conn)] 连接关闭")
	}()

	// 启动一个线程定时发送消息
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case at := <-ticker.C:
				txt, _ := at.MarshalText()
				_, err := conn.Write(txt)
				t.Logf("[CLI virtual conn] 发送消息错误：%v", err)
			case <-signals:
				return
			}
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			t.Logf("[CLI virtual conn] 读取发生错误：%v", err)
			break
		}
		t.Logf("[CLI (virtual conn)] 虚拟连接收到: %s", buf[:n])
	}
}
