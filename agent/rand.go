package agent

import (
	crand "crypto/rand"

	"encoding/binary"
)

// source: https://yourbasic.org/golang/crypto-rand-int/
type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}

func (s cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s cryptoSource) Uint64() (v uint64) {
	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		panic("agent/cryptoSource: " + err.Error())
	}
	return
}
