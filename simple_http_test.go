package spdy_test

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"testing"

	"github.com/dfcfw/spdy"
)

func Test_HTTP_Server(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/tunnel", tunnel)

	if err := http.ListenAndServe(":9998", mux); err != nil {
		t.Fatal(err)
	}
}

func tunnel(w http.ResponseWriter, r *http.Request) {
	// 约定 CONNECT 方法
	if r.Method != http.MethodConnect {
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Println("不是 http.Hijack")
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("Hijack 发生错误：%v", err)
		return
	}

	res := &http.Response{
		Status:     http.StatusText(http.StatusSwitchingProtocols),
		StatusCode: http.StatusSwitchingProtocols,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Request:    r,
	}
	if err = res.Write(conn); err != nil {
		log.Printf("回写失败：%v", err)
		return
	}

	// TODO: 创建多路复用
	spdy.Server(conn)
}

func Test_HTTP_Client(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:9998")
	if err != nil {
		t.Error(err)
		return
	}

	req, _ := http.NewRequest(http.MethodConnect, "http://127.0.0.1:9998/tunnel", nil)
	if err = req.Write(conn); err != nil {
		t.Fatal(err)
	}

	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusSwitchingProtocols {
		t.Fatal("协议升级协商失败")
	}

	// TODO: 创建多路复用
	//mux := spdy.Client(conn)
	//dialFn := func(ctx context.Context, network, addr string) (net.Conn, error) {
	//	if network == "tcp" && addr == "em.com:80" {
	//		return mux.Dial()
	//	}
	//	return nil, &net.AddrError{Addr: addr}
	//}
	//
	//transport := &http.Transport{
	//	DialContext: dialFn,
	//}
	//
	//cli := &http.Client{
	//	Transport: transport,
	//}

}
