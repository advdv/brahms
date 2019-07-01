package agent

import (
	"context"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/advanderveer/brahms"
	httpt "github.com/advanderveer/brahms/transport/http"
)

// Agent participates in a brahm gossip network
type Agent struct {
	rnd       *rand.Rand
	logs      *log.Logger
	self      *brahms.Node
	core      *brahms.Core
	handler   http.Handler
	transport brahms.Transport
	listener  net.Listener
	server    *http.Server
	params    brahms.P

	done chan struct{}

	timeouts struct {
		validate     time.Duration
		update       time.Duration
		invalidation time.Duration
	}
}

// New initializes the agent
func New(logw io.Writer, cfg *Config) (a *Agent, err error) {
	a = &Agent{
		logs:   log.New(logw, "agent/agent: ", 0),
		params: cfg.Params,
		done:   make(chan struct{}),
		rnd:    rand.New(cryptoSource{}),
	}

	a.timeouts.validate = cfg.ValidateTimeout
	a.timeouts.update = cfg.UpdateTimeout
	a.timeouts.invalidation = cfg.InvalidationTimeout

	a.listener, err = net.Listen("tcp", cfg.ListenAddr.String()+":"+strconv.Itoa(int(cfg.ListenPort)))
	if err != nil {
		return nil, Err{err, "listen"}
	}

	a.self = brahms.N(cfg.AdvertiseAddr.String(), cfg.AdvertisePort)
	if a.self.IP == nil {
		a.self.IP = net.ParseIP(a.listener.Addr().(*net.TCPAddr).IP.String())
	}

	if a.self.Port == 0 {
		a.self.Port = uint16(a.listener.Addr().(*net.TCPAddr).Port)
	}

	a.transport = httpt.New(logw)
	return
}

// Self returns info about this agent as a node in the network
func (a *Agent) Self() brahms.Node {
	return *a.self
}

// Join the network and starts the protocol
func (a *Agent) Join(v brahms.View) {
	a.core = brahms.NewCore(a.rnd, a.self, v, a.params, a.transport, a.timeouts.invalidation)
	a.handler = httpt.NewHandler(a.core)
	a.server = &http.Server{
		Handler:      a.handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// start serving http requests
	go func() {
		err := a.server.Serve(a.listener)
		if err != nil && err != http.ErrServerClosed {
			a.logs.Printf("failed to serve http: %v", err)
		}

		close(a.done)
	}()

	// start the protocol loop
	go func() {
		for {
			a.core.UpdateView(a.timeouts.update)
			a.core.ValidateSample(a.timeouts.validate)

			select {
			case <-a.done:
				a.done <- struct{}{}
				return
			default:
			}
		}
	}()
}

// Shutdown attempts to close the agent gracefully
func (a *Agent) Shutdown(ctx context.Context) (err error) {
	if a.core == nil {
		return a.listener.Close()
	}

	a.core.Deactivate()
	a.done <- struct{}{}
	<-a.done

	err = a.server.Shutdown(ctx)
	if err != nil {
		return Err{err, "shutdown"}
	}

	<-a.done
	return
}
