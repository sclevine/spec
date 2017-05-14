package spec

import (
	"math/rand"
	"strings"
	"testing"
)

type summary struct {
	total   int
	pending int
	focused int
	random  bool
	focus   bool
}

type node struct {
	text  []string
	loc   []int
	seed  int64
	order order
	nest  bool
	pend  bool
	focus bool
	nodes tree
}

func (n *node) parse(f func(*testing.T, G, S)) summary {
	var sum summary

	f(nil, func(text string, f func(), opts ...Option) {
		cfg := options(opts).apply()
		if pend, focus := n.add(text, cfg, tree{}); focus && !pend {
			sum.focus = true
		}
		parent := n
		n = n.last()
		defer func() {
			if n.order == orderRandom {
				sum.random = true
			}
			n.sort()
			n = parent
		}()
		f()
	}, func(text string, _ func(), opts ...Option) {
		cfg := options(opts).apply()
		if cfg.before || cfg.after {
			return
		}
		if pend, focus := n.add(text, cfg, nil); focus && !pend {
			sum.focus = true
			sum.focused++
		} else if pend {
			sum.pending++
		}
		sum.total++
	})
	return sum
}

func (n *node) add(text string, cfg *config, nodes tree) (pend, focus bool) {
	pend = cfg.pend || n.pend
	focus = cfg.focus || n.focus
	n.nodes = append(n.nodes, node{
		text:  append(append([]string(nil), n.text...), text),
		loc:   append(append([]int(nil), n.loc...), len(n.nodes)),
		seed:  n.seed,
		order: cfg.order.from(n.order),
		nest:  cfg.nest || n.nest,
		pend:  pend,
		focus: focus,
		nodes: nodes,
	})
	return pend, focus
}

func (n *node) sort() {
	nodes := n.nodes
	switch n.order {
	case orderRandom:
		r := rand.New(rand.NewSource(n.seed))
		for i := len(nodes) - 1; i > 0; i-- {
			j := r.Intn(i + 1)
			nodes[i], nodes[j] = nodes[j], nodes[i]
		}
	case orderReverse:
		last := len(nodes) - 1
		for i := 0; i < len(nodes)/2; i++ {
			nodes[i], nodes[last-i] = nodes[last-i], nodes[i]
		}
	}
}

func (n *node) last() *node {
	return &n.nodes[len(n.nodes)-1]
}

func (n *node) name(full bool) string {
	if !full {
		return n.text[len(n.text)-1]
	}
	return strings.Join(n.text, "/")
}

func (n node) run(t *testing.T, f func(*testing.T, node)) bool {
	switch {
	case n.nodes == nil:
		return t.Run(n.name(!n.nest), func(t *testing.T) { f(t, n) })
	case n.nest:
		return t.Run(n.name(false), func(t *testing.T) { n.nodes.run(t, f) })
	default:
		return n.nodes.run(t, f)
	}
}

type tree []node

func (ns tree) run(t *testing.T, f func(*testing.T, node)) (ok bool) {
	ok = true
	for _, n := range ns {
		ok = n.run(t, f) && ok
	}
	return ok
}
