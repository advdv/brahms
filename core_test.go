package brahms

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/advanderveer/go-test"
)

func TestMiniNetCore(t *testing.T) {
	rnd := rand.New(rand.NewSource(1))
	prm, _ := NewParams(0.45, 0.45, 0.1, 100, 10)

	//create a mini network with three cores
	tr := NewMemNetTransport()
	c1 := NewCore(rnd, NID{0x01}, NewView(NID{0x02}), prm, tr)
	tr.AddCore(c1)
	c2 := NewCore(rnd, NID{0x02}, NewView(NID{0x03}), prm, tr)
	tr.AddCore(c2)
	c3 := NewCore(rnd, NID{0x03}, NewView(NID{0x01}), prm, tr)
	tr.AddCore(c3)

	// after two iterations we should have a connected graph
	for i := 0; i < 10; i++ {
		c1.UpdateView(time.Millisecond)
		c2.UpdateView(time.Millisecond)
		c3.UpdateView(time.Millisecond)
	}

	// view and sampler should show a connected graph
	test.Equals(t, NewView(NID{0x02}, NID{0x03}), c1.view)
	test.Equals(t, NewView(NID{0x02}, NID{0x03}), c1.sampler.Sample())
	test.Equals(t, NewView(NID{0x01}, NID{0x03}), c2.view)
	test.Equals(t, NewView(NID{0x01}, NID{0x03}), c2.sampler.Sample())
	test.Equals(t, NewView(NID{0x01}, NID{0x02}), c3.view)
	test.Equals(t, NewView(NID{0x01}, NID{0x02}), c3.sampler.Sample())
}

func TestNetworkJoin(t *testing.T) {

	//@TODO test how a member would join an existing network by knowing just one
	//other node in the network. A single push without a pull would not be taken
	//into account currently. Is that a real problem in an actual network?

}

func TestLargeNetwork(t *testing.T) {

	r := rand.New(rand.NewSource(1))
	n := uint64(100)
	q := 10
	m := 1.0
	l := int(math.Round(m * math.Pow(float64(n), 1.0/3)))
	p, _ := NewParams(
		0.45,
		0.45,
		0.1,
		l, l,
	)

	tr := NewMemNetTransport()
	cores := make([]*Core, 0, n)
	for i := uint64(1); i <= n; i++ {
		id := NID{}
		other := NID{}
		binary.LittleEndian.PutUint64(id[:], i)
		binary.LittleEndian.PutUint64(other[:], i+1)

		v := NewView(other)
		if i == n {
			v = NewView(NID{0x01}) //loop back to first node, i.e: ring topology
		}

		c := NewCore(r, id, v, p, tr)
		tr.AddCore(c)
		cores = append(cores, c)
	}

	for i := 0; i < q; i++ {
		if !testing.Short() {
			buf := bytes.NewBuffer(nil)
			draw(t, buf, cores)
			drawPNG(t, buf, fmt.Sprintf(filepath.Join("draws", "network_%d.png"), i))
			fmt.Println("drawing step '", i, "'...")
		}

		for _, c := range cores {
			c.UpdateView(time.Microsecond * 700)
		}
	}

	var tot float64
	for _, c := range cores {
		tot += float64(len(c.view))
	}

	// average connectivity should be about this
	test.Equals(t, 3.43, tot/float64(len(cores)))
}
