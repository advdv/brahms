package brahms

import (
	"math/rand"
	"testing"

	"github.com/advanderveer/go-test"
)

func TestViewSortingAndString(t *testing.T) {
	v1 := NewView(NID{0x02}, NID{0x01}, NID{0x03})
	test.Equals(t, []NID{{0x01}, {0x02}, {0x03}}, v1.Sorted())
	for i := 0; i < 100; i++ {
		test.Equals(t, "{0100, 0200, 0300}", v1.String())
	}
}

func TestViewPicking(t *testing.T) {
	for i := 0; i < 100; i++ {
		r := rand.New(rand.NewSource(2))
		v := NewView(NID{0x02}, NID{0x01}, NID{0x03})
		test.Equals(t, "{0100, 0200, 0300}", v.Pick(r, 100).String())
		test.Equals(t, "{0200}", v.Pick(r, 1).String())
	}
}

func TestViewConcat(t *testing.T) {
	v1 := NewView(NID{0x02}, NID{0x01}, NID{0x03})
	v2 := NewView(NID{0x03}, NID{0x05}, NID{0x04})
	v3 := NewView(NID{0x06})
	v1.Concat(v2, v3)
	test.Equals(t, "{0100, 0200, 0300, 0400, 0500, 0600}", v1.String())
}

func TestViewCopy(t *testing.T) {
	v1 := NewView(NID{0x02}, NID{0x01}, NID{0x03})
	v3 := v1              //no copy
	v2 := v1.Copy()       //copy
	delete(v1, NID{0x02}) //edit references

	test.Equals(t, "{0100, 0300}", v3.String())       //no longer has element
	test.Equals(t, "{0100, 0200, 0300}", v2.String()) //still has all elements
}
