package brahms

import (
	"bytes"
	"math/rand"
	"sort"
	"strings"
)

// View describes a set of node ids
type View map[NID]struct{}

// NewView constructs a new view that is a set of the provided ids
func NewView(ids ...NID) (v View) {
	v = View{}
	for _, id := range ids {
		v[id] = struct{}{}
	}

	return
}

// Sorted returns all ids in the view set ordered in lexic order
func (v View) Sorted() (ids []NID) {
	ids = make([]NID, 0, len(v))
	for id := range v {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i int, j int) bool {
		return bytes.Compare(ids[i][:], ids[j][:]) < 0
	})

	return
}

func (v View) String() string {
	fields := make([]string, len(v))
	for i, id := range v.Sorted() {
		fields[i] = id.String()
	}

	return "{" + strings.Join(fields, ", ") + "}"
}

// Pick at most n random members from the set and return them as a new view
func (v View) Pick(r *rand.Rand, n int) (p View) {
	ids := v.Sorted()
	r.Shuffle(len(ids), func(i int, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})

	p = View{}
	for i := 0; i < n; i++ {
		if i >= len(ids) {
			break
		}

		p[ids[i]] = struct{}{}
	}

	return
}

// Concat views to this view
func (v View) Concat(vs ...View) View {
	for _, vv := range vs {
		for id := range vv {
			v[id] = struct{}{}
		}
	}

	return v
}

// Copy returns a copy of this view
func (v View) Copy() View {
	return View{}.Concat(v)
}
