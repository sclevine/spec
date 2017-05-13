package spec

import "testing"

func parse(f func(*testing.T, G, S), cfg *config) (tree, summary) {
	var sum summary
	last := &node{
		order: cfg.order.from(orderSequential),
		pend:  cfg.pend,
		focus: cfg.focus,
	}

	// add sorting

	f(nil, func(text string, f func(), opts ...Option) {
		cfg := options(opts).apply()
		if pend, focus := last.add(text, cfg, tree{}); focus && !pend {
			sum.focus = true
		}
		prev := last
		last = last.last()
		defer func() { last = prev }()
		f()
	}, func(text string, _ func(), opts ...Option) {
		cfg := options(opts).apply()
		if cfg.before || cfg.after {
			return
		}
		if pend, focus := last.add(text, cfg, nil); focus && !pend {
			sum.focus = true
			sum.focused++
		} else if pend {
			sum.pending++
		}
		sum.total++

	})
	return last.nodes, sum
}

type summary struct {
	total   int
	pending int
	focused int
	focus   bool
}

type node struct {
	text  string
	order order
	pend  bool
	focus bool
	index int
	nodes tree
}

func (n *node) add(text string, cfg *config, nodes tree) (pend, focus bool) {
	pend = cfg.pend || n.pend
	focus = cfg.focus || n.focus
	n.nodes = append(n.nodes, node{
		text:  text,
		order: cfg.order.from(n.order),
		pend:  pend,
		focus: focus,
		index: len(n.nodes),
		nodes: nodes,
	})
	return pend, focus
}

func (n *node) last() *node {
	return &n.nodes[len(n.nodes)-1]
}

type tree []node

func (tr tree) run(t *testing.T, groups []int, f func(*testing.T, []int, node)) bool {
	success := true
	for _, n := range tr {
		n := n
		groups := append([]int(nil), groups...)
		success = success && t.Run(n.text, func(t *testing.T) {
			if n.nodes != nil {
				n.nodes.run(t, append(groups, n.index), f)
			} else {
				f(t, groups, n)
			}
		})
	}
	return success
}
