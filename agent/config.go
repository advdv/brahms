package agent

import (
	"net"
	"time"

	"github.com/advanderveer/brahms"
)

//Config configures the agent
type Config struct {
	ListenAddr net.IP
	ListenPort uint16

	AdvertiseAddr net.IP
	AdvertisePort uint16

	ValidationTimeout time.Duration
	UpdateTimeout     time.Duration

	Params brahms.P
}

// LocalTestConfig returns a sensible default config for local testing
func LocalTestConfig() (cfg *Config) {
	cfg = &Config{
		ListenAddr:        net.ParseIP("127.0.0.1"),
		ListenPort:        0,
		ValidationTimeout: time.Millisecond * 100,
		UpdateTimeout:     time.Millisecond * 200,
	}
	cfg.Params, _ = brahms.NewParams(0.45, 0.45, 0.1, 10, 10)
	return
}
