package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
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

	// start reading messages, dedublicate and prevent message storm
	go func() {
		received := map[[32]byte]struct{}{}

		for {
			msg, err := a.Receive()
			if msg == nil || err != nil {
				break
			}

			h := sha256.Sum256(msg)
			if _, ok := received[h]; ok {
				continue //already received
			}

			fmt.Println("new message, relaying:", msg)
			if len(msg) > 0 {
				if a.Emit(msg, 2, 1, time.Second) {
					received[h] = struct{}{}
				}
			}
		}
	}()

	// start emittin messages from the terminal
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			for {
				scanner.Scan()
				msg := scanner.Bytes()
				if len(msg) > 0 {
					fmt.Println(a.Emit(msg, 1, 1, time.Second))
				}
			}
		}()
	}

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
