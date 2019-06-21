package brahms

import (
	"context"
	"math/rand"
	"time"
)

// Brahms implements the gossip protocol and takes an old view 'v' and returns a
// new view.
func Brahms(self *Node, rnd *rand.Rand, p P, to time.Duration, s *Sampler, tr Transport, pushes <-chan *Node, v View) View {

	// reset push/pull views (line 21)
	push, pull := View{}, View{}

	// perform sends and write results to these channels
	pulls := make(chan View, p.βl1())
	func() {
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()

		// push our own id to peers picked from the current view (line 22)
		for id := range v.Pick(rnd, p.αl1()) {
			go tr.Push(ctx, self, id)
		}

		// send pull requests to peers picked from the current view (line 25)
		for id := range v.Pick(rnd, p.βl1()) {
			go tr.Pull(ctx, pulls, id)
		}

		// wait for time unit to be done, cancels any open pushes/pulls (line 27)
		<-ctx.Done()
	}()

	// drain and consider all nodes pushed to us this time period (line 28)
PUSH_DRAIN:
	for {
		select {
		case n := <-pushes:
			push[n.Hash()] = *n
		default:
			break PUSH_DRAIN
		}
	}

	// drain and consider all nodes we pulled in this time period (line 32)
PULL_DRAIN:
	for {
		select {
		case pv := <-pulls:
			for id, n := range pv {
				if id == self.Hash() {
					continue //ignore ourselves if we appear in a pull
				}

				pull[id] = n
			}
		default:
			break PULL_DRAIN
		}
	}

	// only update our view if the nr of pushed ids was not too high (line 35)
	if len(push) <= p.αl1() && len(push) > 0 && len(pull) > 0 {

		// construct our new view from what we've seen this round (line 36)
		v = push.Pick(rnd, p.αl1()).
			Concat(pull.Pick(rnd, p.βl1())).
			Concat(s.Sample().Pick(rnd, p.γl1()))
	}

	// update the sampler with resuling push/pull (line 37)
	s.Update(push.Concat(pull))

	return v
}
