package brahms_test

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"reflect"
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
		c1.ValidateSample(time.Millisecond)
		c2.UpdateView(time.Millisecond)
		c1.ValidateSample(time.Millisecond)
		c3.UpdateView(time.Millisecond)
		c1.ValidateSample(time.Millisecond)
	}

	// view and sampler should show a connected graph
	test.Equals(t, brahms.NewView(n2, n3), c1.ReadView())
	test.Equals(t, brahms.NewView(n2, n3), c1.Sample())
	test.Equals(t, brahms.NewView(n1, n3), c2.ReadView())
	test.Equals(t, brahms.NewView(n1, n3), c2.Sample())
	test.Equals(t, brahms.NewView(n1, n2), c3.ReadView())
	test.Equals(t, brahms.NewView(n1, n2), c3.Sample())

	// test after deactivation
	test.Equals(t, true, c1.IsActive())
	c1.Deactivate()
	test.Equals(t, false, c1.IsActive())
	test.Equals(t, brahms.NewView(), c1.ReadView())
	test.Equals(t, brahms.NewView(), c1.Sample())
}

func TestLargerNetwork(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	r := rand.New(rand.NewSource(1))
	n := uint16(100)
	q := 100

	td := 15
	d := 0.05
	nd := int(math.Round(float64(n) * d))
	nn := 5

	m := 1.0
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
	var lastSample brahms.View
	deactivated := map[brahms.NID]struct{}{}
	for i := 0; i < q; i++ {

		// if not short test: draw graphs
		if !testing.Short() && (5&i == 0 || i == td || i == td+1) {
			views := make(map[*brahms.Node]brahms.View, len(cores))
			dead := make(map[brahms.NID]struct{})
			joins := make(map[brahms.NID]struct{})
			for i, c := range cores {
				cn := c.Self()
				views[cn] = c.Sample()

				if !c.IsActive() {
					dead[cn.Hash()] = struct{}{}
				}

				if i >= int(n) {
					joins[cn.Hash()] = struct{}{}
				}
			}

			wg.Add(1)
			go func(i int, views map[*brahms.Node]brahms.View) {
				defer wg.Done()

				buf := bytes.NewBuffer(nil)
				draw(t, buf, views, dead, joins)
				drawPNG(t, buf, fmt.Sprintf(filepath.Join("_draws", "network_%d.png"), i))
				fmt.Println("drawing step '", i, "'...")

			}(i, views)
		}

		// move the cores ahead in time
		for _, c := range cores {
			if !c.IsActive() {
				continue
			}

			// run update and validation concurrently
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				c.UpdateView(200 * time.Microsecond)
				wg.Done()
			}()
			go func() {
				c.ValidateSample(1000 * time.Microsecond)
				wg.Done()
			}()
			wg.Wait()
		}

		// after some time turn off some cores, and add new ones
		if i == td {
			for i := 0; i < nd; i++ {
				idx := r.Intn(len(cores))
				cores[idx].Deactivate()
				deactivated[cores[idx].Self().Hash()] = struct{}{}
			}

			// add new cores
			for i := len(cores) + 1; i <= int(n)+nn; i++ {
				self := brahms.N("127.0.0.1", uint16(i))
				other := brahms.N("127.0.0.1", uint16(r.Intn(int(n))))

				c := brahms.NewCore(r, self, brahms.NewView(other), p, tr)
				tr.AddCore(c)
				cores = append(cores, c)
			}
		}

		// after a certain round we expect the sample to change very little
		if i > td+5 {
			s := cores[0].Sample()
			if !reflect.DeepEqual(s, lastSample) {
				diff := s.Diff(lastSample)
				if len(diff) > 2 {
					t.Fatalf("observed a significant sample change at %d, new nodes: %s", i, diff)
				}
			}
		}

		lastSample = cores[0].Sample()
	}

	var tot float64
	for i, c := range cores {
		tot += float64(len(c.ReadView()))

		// check that none of the cores still remember the deactivated cores
		for k, _ := range c.Sample() {
			if _, ok := deactivated[k]; ok {
				t.Fatalf("deactivated core should not be in any other cores sample, instead '%s' was in core %d", k, i)
			}
		}
	}

	wg.Wait() //wait for drawings

	test.Assert(t, tot/float64(len(cores)) >= 3.0, fmt.Sprintf("should be reasonably connected, avg is: %f", tot/float64(len(cores))))
}
