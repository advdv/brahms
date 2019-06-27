package agent_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"

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
	a2, err := agent.New(os.Stderr, cfg1)
	test.Ok(t, err)

	self2 := a2.Self()
	test.Equals(t, net.IP{127, 0, 0, 1}, self2.IP)
	test.Assert(t, self2.Port > 0, "should have defaulted to the listening port")

	a2.Join(brahms.NewView())

	_, err = http.Get(fmt.Sprintf("http://127.0.0.1:%d/probe", self2.Port))
	test.Ok(t, err)

	test.Ok(t, a2.Shutdown(context.Background()))
}
