package brahms

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/advanderveer/go-test"
)

func TestViewInit(t *testing.T) {
	n1 := N("127.0.0.1", 1)
	n2 := N("127.0.0.1", 2)

	v1 := NewView(n1, n2)

	n3 := v1.Read(n1.Hash())
	test.Assert(t, fmt.Sprintf("%p", n1) != fmt.Sprintf("%p", n3), "should be different mem locations")
	test.Equals(t, n1, n3) //but same content
	test.Equals(t, (*Node)(nil), v1.Read(NID{}))
}

func TestViewSortingAndString(t *testing.T) {
	n1 := N("127.0.0.1", 1)
	n2 := N("127.0.0.1", 2)
	n3 := N("127.0.0.1", 3)
	v1 := NewView(n1, n2, n3)

	for i := 0; i < 100; i++ {
		test.Equals(t, []Node{*n2, *n3, *n1}, v1.Sorted())
		test.Equals(t, "{127.0.0.1:2, 127.0.0.1:3, 127.0.0.1:1}", v1.String())
	}
}

func TestViewPicking(t *testing.T) {
	n1 := N("127.0.0.1", 1)
	n2 := N("127.0.0.1", 2)
	n3 := N("127.0.0.1", 3)

	for i := 0; i < 100; i++ {
		r := rand.New(rand.NewSource(2))
		v := NewView(n1, n2, n3)

		test.Equals(t, v, v.Pick(r, 100))
		test.Equals(t, NewView(n3), v.Pick(r, 1))
	}
}

func TestViewConcat(t *testing.T) {
	n1 := N("127.0.0.1", 1)
	n2 := N("127.0.0.1", 2)
	n3 := N("127.0.0.1", 3)
	n4 := N("127.0.0.1", 4)
	n5 := N("127.0.0.1", 5)
	n6 := N("127.0.0.1", 6)

	v1 := NewView(n2, n1, n3)
	v2 := NewView(n3, n5, n4)
	v3 := NewView(n6)
	v1.Concat(v2, v3)

	test.Equals(t, NewView(n1, n2, n3, n4, n5, n6), v1)
}

func TestViewCopy(t *testing.T) {
	n1 := N("127.0.0.1", 1)
	n2 := N("127.0.0.1", 2)
	n3 := N("127.0.0.1", 3)

	v1 := NewView(n2, n1, n3)
	v3 := v1                                //no copy
	v2 := v1.Copy()                         //copy
	delete(v1, n2.Hash())                   //edit references
	test.Equals(t, NewView(n1, n3), v3)     //no longer has element
	test.Equals(t, NewView(n1, n2, n3), v2) //still has all elements
}
