package spec

import "testing"

func parse(f func(*testing.T, G, S), opts ...Option) tree {
	cfg := options(opts).apply()
	last := &node{
		order: cfg.order.from(orderSequential),
		pend: cfg.pend,
		focus: cfg.focus,
	}

	f(nil, func(text string, f func(), opts ...Option) {
		cfg := options(opts).apply()
		last.nodes = append(last.nodes, node{
			text:  text,
			order: cfg.order,
			pend:  cfg.pend,
			focus: cfg.focus,
			index: len(last.nodes),
			nodes: tree{},
		})
		prevLast := last
		last = &last.nodes[len(last.nodes)-1]
		defer func() { last = prevLast }()
		f()
	}, func(text string, _ func(), opts ...Option) {
		cfg := options(opts).apply()
		if cfg.before || cfg.after {
			return
		}
		last.nodes = append(last.nodes, node{
			text:  text,
			order: cfg.order,
			pend:  cfg.pend,
			focus: cfg.focus,
			index: len(last.nodes),
		})
	})
	return last.nodes
}

type node struct {
	text  string
	order order
	pend  bool
	focus bool
	index int
	nodes tree
}

type tree []node

func (tr tree) run(t *testing.T, groups []int, f func(*testing.T, []int, node)) bool {
	success := true
	for _, n := range tr {
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
