package httpt_test

import (
	"bytes"
	"context"
	"net"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/advanderveer/brahms"
	httpt "github.com/advanderveer/brahms/transport/http"
	"github.com/advanderveer/go-test"
)

var _ brahms.Transport = &httpt.Transport{}

func TestTransportRequest(t *testing.T) {
	b := &mockBrahms{}
	s := httptest.NewServer(httpt.NewHandler(b))

	defer s.Close()
	host, ports, _ := net.SplitHostPort(s.Listener.Addr().String())
	port, _ := strconv.Atoi(ports)

	buf := bytes.NewBuffer(nil)
	tr := httpt.New(buf)

	t.Run("invalid request", func(t *testing.T) {
		err := tr.Request(context.Background(), "GET ", *brahms.N(host, uint16(port)), "/probe", nil, nil)
		test.Equals(t, "request_creation", err.(httpt.TransportErr).Op)
	})

	t.Run("request execution", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
		defer cancel()

		err := tr.Request(ctx, "GET", *brahms.N(host, uint16(port)), "/probe", nil, nil)
		test.Equals(t, "request_execution", err.(httpt.TransportErr).Op)
	})

	t.Run("request execution", func(t *testing.T) {
		err := tr.Request(context.Background(), "GET", *brahms.N(host, uint16(port)), "/def", nil, map[string]interface{}{})
		test.Equals(t, "response_decoding", err.(httpt.TransportErr).Op)
	})

	t.Run("request execution", func(t *testing.T) {
		tr.RequestOrLog(context.Background(), "GET", *brahms.N(host, uint16(port)), "/def", nil, map[string]interface{}{})
		test.Assert(t, strings.Contains(buf.String(), "failed to perform request"), "should have logged failure")
	})
}

func TestTransport(t *testing.T) {
	b := &mockBrahms{}
	s := httptest.NewServer(httpt.NewHandler(b))

	defer s.Close()
	host, ports, _ := net.SplitHostPort(s.Listener.Addr().String())
	port, _ := strconv.Atoi(ports)
	tr := httpt.New(os.Stderr)

	t.Run("probe, push", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		c := make(chan brahms.NID, 1)
		tr.Probe(ctx, c, brahms.NID{0x01}, *brahms.N(host, uint16(port)))
		tr.Push(ctx, *brahms.N("127.0.0.1", 9090), *brahms.N(host, uint16(port)))

		select {
		case <-time.After(time.Second):
			t.Fatal("took too long")
		case i := <-c:
			test.Equals(t, brahms.NID{0x01}, i)
			test.Equals(t, 1, len(b.pushes))
			test.Equals(t, net.ParseIP("127.0.0.1"), b.pushes[0].IP)
			test.Equals(t, uint16(9090), b.pushes[0].Port)
		}
	})

	t.Run("pull", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		c := make(chan brahms.View, 1)
		tr.Pull(ctx, c, *brahms.N(host, uint16(port)))

		select {
		case <-time.After(time.Second):
			t.Fatal("took too long")
		case v := <-c:
			test.Equals(t, 1, len(v))
			for _, n := range v {
				test.Equals(t, "127.0.0.1", n.IP.String())
				test.Equals(t, uint16(8080), n.Port)
			}
		}
	})

}
