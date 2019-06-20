package brahms

import (
	"context"
	"math/rand"
	"time"
)

// Brahms implements the gossip protocol and takes an old view 'v' and returns a
// new view.
func Brahms(self NID, rnd *rand.Rand, p P, to time.Duration, s *Sampler, tr Transport, pushes <-chan NID, v View) View {

	// reset push/pull views (line 21)
	push, pull := View{}, View{}

	// perform sends and write results to these channels
	cpull := make(chan View, p.βl1())
	func() {
		ctx, cancel := context.WithTimeout(context.Background(), to)
		defer cancel()

		// push our own id to peers picked from the current view (line 22)
		for id := range v.Pick(rnd, p.αl1()) {
			go tr.Push(ctx, self, id)
		}

		// send pull requests to peers picked from the current view (line 25)
		for id := range v.Pick(rnd, p.βl1()) {
			go tr.Pull(ctx, cpull, id)
		}

		// wait for time unit to be done, cancels any open pushes/pulls (line 27)
		<-ctx.Done()
		close(cpull)
	}()

	// drain the buffer of all ids pushed to us (line 28)
DRAIN:
	for {
		select {
		case id := <-pushes:
			push[id] = struct{}{}
		default:
			break DRAIN
		}
	}

	// add all peers that we received as replies from our pull requests (line 32)
	for pv := range cpull {
		for id := range pv {

			//NOTE: we divert from the paper by ignoring any pulls
			if id == self {
				continue
			}

			pull[id] = struct{}{}
		}
	}

	// only update our view if the nr of pushed ids was not too high (line 35)
	// NOTE: we divert from the paper here. We're happy to update if either pull
	// or push yielded us some nodes not necessarily both.
	if len(push) <= p.αl1() && (len(push) > 0 || len(pull) > 0) {

		// construct our new view from what we've seen this round (line 36)
		v = push.Pick(rnd, p.αl1()).
			Concat(pull.Pick(rnd, p.βl1())).
			Concat(s.Sample().Pick(rnd, p.γl1()))
	}

	// update the sampler with resuling push/pull (line 37)
	s.Update(push.Concat(pull))

	return v
}
