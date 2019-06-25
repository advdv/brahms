package brahms

import (
	"bytes"
	"math/rand"
	"sort"
	"strings"
)

// View describes a set of node ids
type View map[NID]Node

// NewView constructs a new view that is a set of a copy of the provided node info.
func NewView(ns ...*Node) (v View) {
	v = View{}
	for _, n := range ns {
		v[n.Hash()] = *n
	}

	return
}

// Read a reference to a copy of the node with the provided id
func (v View) Read(id NID) *Node {
	n, ok := v[id]
	if !ok {
		return nil
	}

	return &n
}

// Sorted returns all ids in the view set ordered in lexic order
func (v View) Sorted() (ns []Node) {
	ids := make([]NID, 0, len(v))
	for id := range v {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i int, j int) bool {
		return bytes.Compare(ids[i][:], ids[j][:]) < 0
	})

	ns = make([]Node, len(ids))
	for i, id := range ids {
		ns[i] = v[id]
	}

	return
}

func (v View) String() string {
	fields := make([]string, len(v))
	for i, n := range v.Sorted() {
		fields[i] = n.String()
	}

	return "{" + strings.Join(fields, ", ") + "}"
}

// Pick at most n random members from the set and return them as a new view
func (v View) Pick(r *rand.Rand, n int) (p View) {
	ns := v.Sorted()
	r.Shuffle(len(ns), func(i int, j int) {
		ns[i], ns[j] = ns[j], ns[i]
	})

	p = View{}
	for i := 0; i < n; i++ {
		if i >= len(ns) {
			break
		}

		p[ns[i].Hash()] = ns[i]
	}

	return
}

// Concat views to this view and return it
func (v View) Concat(vs ...View) View {
	for _, vv := range vs {
		for id, n := range vv {
			v[id] = n
		}
	}

	return v
}

// Copy returns a copy of this view
func (v View) Copy() View {
	return View{}.Concat(v)
}

// Inter returns the intersection between two views
func (v View) Inter(other View) (sect View) {
	sect = View{}
	for id, n := range v {
		if _, ok := other[id]; ok {
			sect[id] = n
		}
	}

	return
}

// Diff returns all elements that are in this view but not in the other view
func (v View) Diff(other View) (d View) {
	d = View{}
	for id, n := range v {
		if _, ok := other[id]; !ok {
			d[id] = n
		}
	}

	return
}
