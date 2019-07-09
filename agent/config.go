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

	ValidateTimeout     time.Duration
	UpdateTimeout       time.Duration
	InvalidationTimeout time.Duration
	ReceiveTimeout      time.Duration

	Params brahms.P
}

// LocalTestConfig returns a sensible default config for local testing
func LocalTestConfig() (cfg *Config) {
	cfg = &Config{
		ListenAddr:          net.ParseIP("127.0.0.1"),
		ListenPort:          0,
		ValidateTimeout:     time.Millisecond * 100,
		UpdateTimeout:       time.Millisecond * 200,
		InvalidationTimeout: time.Second * 5,
		ReceiveTimeout:      time.Second,
	}

	cfg.Params, _ = brahms.NewParams(0.45, 0.45, 0.1, 10, 10, 2)
	return
}
