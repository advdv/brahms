package agent

import (
	"context"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync"
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
	handler   *httpt.Handler
	transport brahms.Transport
	listener  net.Listener
	server    *http.Server
	params    brahms.P

	done chan struct{}

	timeouts struct {
		validate     time.Duration
		update       time.Duration
		invalidation time.Duration
		receive      time.Duration
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
	a.timeouts.receive = cfg.ReceiveTimeout

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

// Emit dissemates the message to N peers and succeeds unless less than m
// peers responded with success
func (a *Agent) Emit(msg []byte, n, m int, to time.Duration) (ok bool) {
	peers := a.core.Sample().Pick(a.rnd, n)

	// emit to all peers, done indicates either the ctx expired or all responded in time
	emits := make(chan brahms.NID, len(peers))
	done := make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()

		var wg sync.WaitGroup
		for id, p := range peers {
			wg.Add(1)

			// run transport request in separate routines
			go func(id brahms.NID, p brahms.Node) { a.transport.Emit(ctx, emits, id, msg, p); wg.Done() }(id, p)
		}

		//wait for them to finish, context will cancel if it takes too long
		wg.Wait()
		close(done)
	}()
	<-done

	// drain the ok's we got at this point
	oks := map[brahms.NID]struct{}{}
DRAIN:
	for {
		select {
		case id := <-emits:
			oks[id] = struct{}{}
		default:
			break DRAIN
		}
	}

	if len(oks) < m {
		return false
	}

	return true
}

// Receive will block until a new message can be read from the network
func (a *Agent) Receive() (msg []byte, err error) {
	if a.handler == nil {
		// @TODO if we call receive when there is no handler we need to block a bit
		// to not exhaust the cpu in an uncostrained for loop. In reality we would
		// like to just initiate the handler when we initiate the agent
		time.Sleep(time.Millisecond)
		return nil, Err{errors.New("uninitialized handler"), "receive"}
	}

	msg = <-a.handler.C
	if msg == nil {
		//@TODO allow handler shutdown to actually trigger this
		return nil, io.EOF
	}

	return
}

// Join the network and starts the protocol
func (a *Agent) Join(v brahms.View) {
	a.core = brahms.NewCore(a.rnd, a.self, v, a.params, a.transport, a.timeouts.invalidation)
	a.handler = httpt.NewHandler(a.core, 0, a.timeouts.receive)
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
