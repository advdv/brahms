package brahms

import (
	"net"
	"testing"

	"github.com/advanderveer/go-test"
)

func TestNIDString(t *testing.T) {
	test.Equals(t, true, NID{}.IsNil())
	test.Equals(t, "0101", NID{0x01, 0x01}.String())
	test.Equals(t, byte(0x01), NID{0x01}.Bytes()[0])
	test.Equals(t, 32, len(NID{0x01}.Bytes()))
}

func TestNodeCreation(t *testing.T) {
	n1 := N("127.0.0.1", 10000)
	test.Equals(t, "127.0.0.1", n1.IP.String())
	test.Equals(t, uint16(10000), n1.Port)
}

func TestNodeHashing(t *testing.T) {
	n := Node{}
	test.Equals(t, true, n.IsZero())
	test.Equals(t, "96a2", n.Hash().String())

	n = Node{}
	n.IP = net.ParseIP("127.0.0.1")
	test.Equals(t, "53e7", n.Hash().String())

	n = Node{}
	n.Port = 1
	test.Equals(t, "b413", n.Hash().String())
	test.Equals(t, false, n.IsZero())
}

var rid NID

func BenchmarkNodeHashing(b *testing.B) {
	var id NID

	n := N("127.0.0.1", 11000)
	for i := 0; i < b.N; i++ {
		id = n.Hash() //~253ns
	}

	rid = id
}
