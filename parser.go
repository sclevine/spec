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
	name  []string
	loc   []int
	seed  int64
	order order
	scope scope
	nest  nest
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
			n.flatten()
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
	// TODO: cfg validation logic here

	name := n.name
	if n.nest == nestOn {
		name = nil
	}
	pend = cfg.pend || n.pend
	focus = cfg.focus || n.focus
	n.nodes = append(n.nodes, node{
		name:  append(append([]string(nil), name...), text),
		loc:   append(append([]int(nil), n.loc...), len(n.nodes)),
		seed:  n.seed,
		order: cfg.order.or(n.order),
		scope: cfg.scope.or(n.scope),
		nest:  cfg.nest.or(n.nest),
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

func (n *node) flatten() {
	nodes := n.nodes
	switch n.scope {
	case scopeGlobal:
		var flat tree
		for _, child := range nodes {
			if child.nodes == nil || child.scope == scopeLocal {
				flat = append(flat, child)
			} else {
				flat = append(flat, child.nodes...)
			}
		}
		n.nodes = flat
	}
}

func (n *node) last() *node {
	return &n.nodes[len(n.nodes)-1]
}

func (n node) run(t *testing.T, f func(*testing.T, node)) bool {
	name := strings.Join(n.name, "/")
	switch {
	case n.nodes == nil:
		return t.Run(name, func(t *testing.T) { f(t, n) })
	case n.nest == nestOn:
		return t.Run(name, func(t *testing.T) { n.nodes.run(t, f) })
	default:
		return n.nodes.run(t, f)
	}
}

type tree []node

func (ns tree) run(t *testing.T, f func(*testing.T, node)) bool {
	ok := true
	for _, n := range ns {
		ok = n.run(t, f) && ok
	}
	return ok
}
