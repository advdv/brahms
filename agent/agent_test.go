package agent_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/advanderveer/brahms"
	"github.com/advanderveer/brahms/agent"
	"github.com/advanderveer/go-test"
)

func TestAgentInit(t *testing.T) {
	cfg1 := agent.LocalTestConfig()
	cfg1.ListenAddr = nil
	_, err := agent.New(os.Stderr, cfg1)
	test.Equals(t, "listen", err.(agent.Err).Op)

	cfg1.ListenAddr = net.IP{127, 0, 0, 1}
	a, err := agent.New(os.Stderr, cfg1)
	test.Ok(t, err)

	// should have relevant self info
	self2 := a.Self()
	test.Equals(t, net.ParseIP("127.0.0.1"), self2.IP)
	test.Assert(t, self2.Port > 0, "should have defaulted to the listening port")

	// should be able to shutdown before join was called, and then start again
	test.Ok(t, a.Shutdown(context.Background()))
	a, err = agent.New(os.Stderr, cfg1)
	test.Ok(t, err)
	self2 = a.Self()

	// then start an enmpty group
	a.Join(brahms.NewView())

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/probe", self2.Port))
	test.Ok(t, err)
	test.Equals(t, http.StatusOK, resp.StatusCode)

	test.Ok(t, a.Shutdown(context.Background()))
}

func TestSmallAgentNetwork(t *testing.T) {
	n, q, m := 5, 3, 3

	done := make(chan struct{}, q)
	agents := make([]*agent.Agent, 0, n)
	for i := 0; i < n; i++ {
		cfg := agent.LocalTestConfig()
		cfg.ReceiveTimeout = time.Millisecond * 40

		a, err := agent.New(os.Stderr, cfg)
		test.Ok(t, err)

		agents = append(agents, a)
		go func(a *agent.Agent) {
			for {
				msg, err := a.Receive()
				if msg == nil || err != nil {
					continue
				}

				test.Equals(t, []byte("foo"), msg)
				done <- struct{}{}
			}
		}(a)

	}

	for _, a := range agents {
		first := agents[0].Self()
		a.Join(brahms.NewView(&first))
	}

	test.Equals(t, false, agents[0].Emit([]byte("foo"), q, m, time.Millisecond*100))
	time.Sleep(time.Millisecond * 400)
	test.Equals(t, true, agents[0].Emit([]byte("foo"), q, m, time.Second))

	<-done
	<-done
	<-done
	<-done
}
