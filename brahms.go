package brahms

import (
	"context"
	"math/rand"
	"time"
)

// Transport describes how a node communicates with its peers
type Transport interface {
	Push(ctx context.Context, self Node, to Node)
	Pull(ctx context.Context, c chan<- View, from Node)
	Prober
}

// Brahms implements the gossip protocol and takes an old view 'v' and returns a
// new view.
func Brahms(self *Node, rnd *rand.Rand, p P, to time.Duration, s *Sampler, tr Transport, pushes <-chan Node, v View) View {

	// reset push/pull views (line 21)
	push, pull := View{}, View{}

	// perform sends and write results to these channels
	pulls := make(chan View, p.L1β())
	func() {
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()

		// push our own id to peers picked from the current view (line 22)
		for _, n := range v.Pick(rnd, p.L1α()) {
			go tr.Push(ctx, *self, n)
		}

		// send pull requests to peers picked from the current view (line 25)
		for _, n := range v.Pick(rnd, p.L1β()) {
			go tr.Pull(ctx, pulls, n)
		}

		// wait for time unit to be done, cancels any open pushes/pulls (line 27)
		<-ctx.Done()
	}()

	// drain and consider all nodes pushed to us this time period (line 28)
PUSH_DRAIN:
	for {
		select {
		case n := <-pushes:
			push[n.Hash()] = n
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
	if len(push) <= p.L1α() && len(push) > 0 && len(pull) > 0 {

		// construct our new view from what we've seen this round (line 36)
		v = push.Pick(rnd, p.L1α()).
			Concat(pull.Pick(rnd, p.L1β())).
			Concat(s.Sample().Pick(rnd, p.L1γ()))
	}

	// update the sampler with resuling push/pull (line 37)
	s.Update(push.Concat(pull))

	return v
}
