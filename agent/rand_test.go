package agent

import (
	"math/rand"
	"testing"
)

var _ rand.Source = cryptoSource{}

func TestCRandSource(t *testing.T) {
	src := cryptoSource{}
	var v int64
	for i := 0; i < 100; i++ {
		nv := src.Int63()
		if nv == v {
			t.Fatal("random value was the same as last value")
		}
		v = nv
	}

	var w uint64
	for i := 0; i < 100; i++ {
		nw := src.Uint64()
		if nw == w {
			t.Fatal("random value was the same as last value")
		}
		w = nw
	}
}
