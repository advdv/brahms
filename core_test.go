package brahms_test

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/advanderveer/brahms"
	"github.com/advanderveer/brahms/transport"
	"github.com/advanderveer/go-test"
)

func TestMiniNetCore(t *testing.T) {
	n1 := brahms.N("127.0.0.1", 1)
	n2 := brahms.N("127.0.0.1", 2)
	n3 := brahms.N("127.0.0.1", 3)

	rnd := rand.New(rand.NewSource(1))
	prm, _ := brahms.NewParams(0.45, 0.45, 0.1, 100, 10)

	//create a mini network with three cores
	tr := transport.NewMemNetTransport()
	c1 := brahms.NewCore(rnd, n1, brahms.NewView(n2), prm, tr)
	tr.AddCore(c1)
	c2 := brahms.NewCore(rnd, n2, brahms.NewView(n3), prm, tr)
	tr.AddCore(c2)
	c3 := brahms.NewCore(rnd, n3, brahms.NewView(n1), prm, tr)
	tr.AddCore(c3)

	// after two iterations we should have a connected graph
	for i := 0; i < 10; i++ {
		c1.UpdateView(time.Millisecond)
		c2.UpdateView(time.Millisecond)
		c3.UpdateView(time.Millisecond)
	}

	// view and sampler should show a connected graph
	test.Equals(t, brahms.NewView(n2, n3), c1.View())
	test.Equals(t, brahms.NewView(n2, n3), c1.Sample())
	test.Equals(t, brahms.NewView(n1, n3), c2.View())
	test.Equals(t, brahms.NewView(n1, n3), c2.Sample())
	test.Equals(t, brahms.NewView(n1, n2), c3.View())
	test.Equals(t, brahms.NewView(n1, n2), c3.Sample())
}

func TestNetworkJoin(t *testing.T) {
	//@TODO test how a member would join an existing network by knowing just one
	//other node in the network. A single push without a pull would not be taken
	//into account currently. Is that a real problem in an actual network?
}

func TestLargeNetwork(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	n := uint16(100)
	q := 40

	td := 20
	d := 0.05
	nd := int(math.Round(float64(n) * d))

	m := 2.0
	l := int(math.Round(m * math.Pow(float64(n), 1.0/3)))
	p, _ := brahms.NewParams(
		0.45,
		0.45,
		0.1,
		l, l,
	)

	tr := transport.NewMemNetTransport()
	cores := make([]*brahms.Core, 0, n)
	for i := uint16(1); i <= n; i++ {
		self := brahms.N("127.0.0.1", i)
		other := brahms.N("127.0.0.1", i+1)
		if i == n {
			other = brahms.N("127.0.0.1", 1)
		}

		c := brahms.NewCore(r, self, brahms.NewView(other), p, tr)
		tr.AddCore(c)
		cores = append(cores, c)
	}

	var wg sync.WaitGroup
	for i := 0; i < q; i++ {

		// if not short test: draw graphs
		if !testing.Short() && (5&i == 0 || i == td || i == td+1) {
			views := make(map[*brahms.Node]brahms.View, len(cores))
			dead := make(map[brahms.NID]struct{})
			for _, c := range cores {
				views[c.Self()] = c.View().Copy()

				if !c.IsAlive() {
					dead[c.Self().Hash()] = struct{}{}
				}
			}

			wg.Add(1)
			go func(i int, views map[*brahms.Node]brahms.View) {
				defer wg.Done()

				buf := bytes.NewBuffer(nil)
				draw(t, buf, views, dead)
				drawPNG(t, buf, fmt.Sprintf(filepath.Join("draws", "network_%d.png"), i))
				fmt.Println("drawing step '", i, "'...")

			}(i, views)
		}

		// move the cores ahead in time
		for _, c := range cores {
			if !c.IsAlive() {
				continue
			}

			c.UpdateView(100 * time.Microsecond)
			c.ValidateSample(100 * time.Microsecond)
		}

		// after some time turn off some cores
		if i == td {
			for i := 0; i < nd; i++ {
				idx := r.Intn(len(cores))
				cores[idx].SetAlive(false)
				cores[idx].ClearView()
			}
		}
	}

	var tot float64
	for _, c := range cores {
		tot += float64(len(c.View()))
	}

	wg.Wait() //wait for drawings

	// @TODO assert that no-one connects to the in-active cores anymore
	// @TODO assert that the rest is still connected
	// test.Assert(t, tot/float64(len(cores)) > 3.1, "should be reasonably connected")
}
