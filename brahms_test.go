package brahms

import (
	"math/rand"
	"testing"
	"time"

	"github.com/advanderveer/go-test"
)

func TestBrahmsNoReply(t *testing.T) {
	n1 := N("127.0.0.1", 1)

	p, _ := NewParams(0.1, 0.7, 0.2, 10, 2)
	r := rand.New(rand.NewSource(1))
	s := NewSampler(r, p.l2())
	self := n1

	p0 := make(chan *Node)
	v0 := NewView(n1)
	tr0 := NewMockTransport()

	v1 := Brahms(self, r, p, time.Millisecond*10, s, tr0, p0, v0)

	//view should be unchanged transport returned nothing
	test.Equals(t, v0, v1)

	//should have pushed our own nid
	test.Equals(t, true, tr0.DidPush(self.Hash()))

	//sample should be empty
	test.Equals(t, View{}, s.Sample())
}

func TestBrahmsWithJustPushes(t *testing.T) {
	n1 := N("127.0.0.1", 1)
	n2 := N("127.0.0.1", 2)
	n3 := N("127.0.0.1", 3)
	n4 := N("127.0.0.1", 4)
	n5 := N("127.0.0.1", 5)

	p, _ := NewParams(0.1, 0.7, 0.2, 10, 2)
	r := rand.New(rand.NewSource(1))
	s := NewSampler(r, p.l2())
	self := n1

	p0 := make(chan *Node, 10)
	p0 <- n2
	v0 := NewView(n1)
	tr0 := NewMockTransport()

	// with just a pull response we do not update the view with just that info
	v1 := Brahms(self, r, p, time.Millisecond*10, s, tr0, p0, v0)
	test.Equals(t, 0, len(p0))
	test.Equals(t, v0, v1)

	// but the pushed id should have been added to the sample
	test.Equals(t, NewView(n2), s.Sample())

	t.Run("with too many pushes", func(t *testing.T) {
		p1 := make(chan *Node, 10)
		p1 <- n3
		p1 <- n4 //with the given params this is too much push

		v2 := Brahms(n5, r, p, time.Millisecond*10, s, tr0, p1, v0)

		//with too many pushes the view shouldn't have changed
		test.Equals(t, v2, v0)
	})
}

func TestBrahmsWithPullsAndPushes(t *testing.T) {
	n1 := N("127.0.0.1", 1)
	n2 := N("127.0.0.1", 2)
	n3 := N("127.0.0.1", 3)
	n4 := N("127.0.0.1", 4)

	p, _ := NewParams(0.1, 0.7, 0.2, 10, 4)
	r := rand.New(rand.NewSource(1))
	s := NewSampler(r, p.l2())
	self := n1
	other := n2

	v0 := NewView(other)
	p0 := make(chan *Node, 10)
	p0 <- n4
	tr0 := NewMockTransport()
	tr0.SetPull(other.Hash(), NewView(n3, n3, self))

	//with both pushes and pulls the view should get updated
	v1 := Brahms(self, r, p, time.Millisecond*10, s, tr0, p0, v0)

	test.Equals(t, NewView(n3, n4), v1)
	test.Equals(t, NewView(n3, n4), s.Sample())
}
