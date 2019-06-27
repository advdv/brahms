package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/advanderveer/brahms"
	"github.com/advanderveer/brahms/agent"
)

func main() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, os.Interrupt)

	cfg := agent.LocalTestConfig()
	if os.Getenv("PORT") != "" {
		port, _ := strconv.Atoi(os.Getenv("PORT"))
		cfg.ListenPort = uint16(port)
	}

	cfg.UpdateTimeout = time.Second * 1
	cfg.ValidateTimeout = time.Second * 1

	a, err := agent.New(os.Stderr, cfg)
	if err != nil {
		panic(err)
	}

	v := brahms.NewView()
	if len(os.Args) > 1 {
		host, ports, err := net.SplitHostPort(os.Args[1])
		if err != nil {
			panic("invalid host/port arg")
		}

		port, err := strconv.Atoi(ports)
		if err != nil {
			panic("invalid host/port arg")
		}

		v = brahms.NewView(brahms.N(host, uint16(port)))
	}

	a.Join(v)

	log.Printf("agent started with v0=%s, advertising as: %v", v, a.Self())
	sig := <-sigs
	log.Printf("received %s, shutting down gracefully", sig)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = a.Shutdown(ctx)
	if err != nil {
		panic(err)
	}

}
