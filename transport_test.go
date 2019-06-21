package brahms

import (
	"context"
	"testing"

	"github.com/advanderveer/go-test"
)

func TestNetCoreTranport(t *testing.T) {
	n1 := N("127.0.0.1", 1)
	n2 := N("127.0.0.1", 2)

	tr := NewMemNetTransport()

	t.Run("push", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		tr.Push(nil, *n1, *n2)
	})

	t.Run("pull", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		tr.Pull(nil, nil, *n2)
	})

	t.Run("probe", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		tr.Probe(nil, nil, 1, *n2)
	})
}

func TestMockTransportProbe(t *testing.T) {
	tr := NewMockTransport()
	c := make(chan int, 1)
	tr.Probe(context.Background(), c, 1, Node{})
	test.Equals(t, 1, <-c)
}
