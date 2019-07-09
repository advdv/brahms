package brahms_test

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/advanderveer/brahms"
	"github.com/advanderveer/brahms/transport"
	"github.com/advanderveer/go-test"
)

func TestBrahmsNoReply(t *testing.T) {
	n1 := brahms.N("127.0.0.1", 1)
	pr := proberFunc(func(ctx context.Context, c chan<- brahms.NID, i brahms.NID, n brahms.Node) {})

	p, _ := brahms.NewParams(0.1, 0.7, 0.2, 10, 2, 2)
	r := rand.New(rand.NewSource(0))
	s := brahms.NewSampler(r, p.L2(), pr, time.Second)
	self := n1

	p0 := make(chan brahms.Node)
	v0 := brahms.NewView(n1)
	tr0 := transport.NewMockTransport()

	v1 := brahms.Brahms(self, r, p, time.Millisecond*10, s, tr0, p0, v0)

	//view should be reset if transport returned nothing and sampler is empty
	test.Equals(t, brahms.NewView(), v1)

	//should have pushed our own nid
	test.Equals(t, true, tr0.DidPush(self.Hash()))

	//sample should be empty
	test.Equals(t, brahms.View{}, s.Sample())
}

func TestBrahmsWithJustPushes(t *testing.T) {
	n1 := brahms.N("127.0.0.1", 1)
	n2 := brahms.N("127.0.0.1", 2)
	n3 := brahms.N("127.0.0.1", 3)
	n4 := brahms.N("127.0.0.1", 4)
	n5 := brahms.N("127.0.0.1", 5)

	p, _ := brahms.NewParams(0.1, 0.7, 0.2, 10, 2, 2)
	r := rand.New(rand.NewSource(1))
	pr := proberFunc(func(ctx context.Context, c chan<- brahms.NID, id brahms.NID, n brahms.Node) {})
	s := brahms.NewSampler(r, p.L2(), pr, time.Second)
	self := n1

	p0 := make(chan brahms.Node, 10)
	p0 <- *n2
	p0 <- *self //pushing ourself should be ignored
	v0 := brahms.NewView(n1)
	tr0 := transport.NewMockTransport()

	// with just a pull response we do not update the view with just that info
	v1 := brahms.Brahms(self, r, p, time.Millisecond*10, s, tr0, p0, v0)
	test.Equals(t, 0, len(p0))
	test.Equals(t, brahms.NewView(n2), v1)

	// but the pushed id should have been added to the sample
	test.Equals(t, brahms.NewView(n2), s.Sample())

	t.Run("with too many pushes", func(t *testing.T) {
		p1 := make(chan brahms.Node, 10)
		p1 <- *n3
		p1 <- *n4 //with the given params this is too much push

		v2 := brahms.Brahms(n5, r, p, time.Millisecond*10, s, tr0, p1, v0)

		//with too many pushes the view shouldn't have changed
		test.Equals(t, v2, v0)
	})
}

func TestBrahmsWithPullsAndPushes(t *testing.T) {
	n1 := brahms.N("127.0.0.1", 1)
	n2 := brahms.N("127.0.0.1", 2)
	n3 := brahms.N("127.0.0.1", 3)
	n4 := brahms.N("127.0.0.1", 4)
	n5 := brahms.N("127.0.0.1", 5)

	p, _ := brahms.NewParams(0.1, 0.7, 0.2, 10, 4, 2)
	r := rand.New(rand.NewSource(1))
	pr := proberFunc(func(ctx context.Context, c chan<- brahms.NID, i brahms.NID, n brahms.Node) {})
	s := brahms.NewSampler(r, p.L2(), pr, time.Second)

	// sample n5, then probe it to recently invalidate it
	s.Update(brahms.NewView(n5))
	s.Validate(r, 4, time.Millisecond*5)

	self := n1
	other := n2
	v0 := brahms.NewView(other)
	p0 := make(chan brahms.Node, 10)
	p0 <- *n4
	tr0 := transport.NewMockTransport()
	tr0.SetPull(other.Hash(), brahms.NewView(n3, n3, n5, self))

	// with both pushes and pulls the view should get updated.
	// should ignore self and n5, the latter because it was recently invalidated
	v1 := brahms.Brahms(self, r, p, time.Millisecond*10, s, tr0, p0, v0)
	test.Equals(t, brahms.NewView(n3, n4), v1)
	test.Equals(t, brahms.NewView(n3, n4), s.Sample())
}
