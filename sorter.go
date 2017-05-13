package spec

import "testing"

func parse(f func(*testing.T, G, S), opts ...Option) tree {
	var nodes *tree
	cfg := options(opts).apply()

	f(nil, func(text string, f func(), opts ...Option) {
		cfg := options(opts).apply()
		groupNode := node{
			text:  text,
			order: cfg.order,
			pend:  cfg.pend,
			focus: cfg.focus,
			index: len(*nodes),
			nodes: tree{},
		}
		*nodes = append(*nodes, groupNode)
		prevNodes := nodes
		nodes = &groupNode.nodes
		defer func() { nodes = prevNodes }()
		f()
	}, func(text string, _ func(), opts ...Option) {
		cfg := options(opts).apply()
		if cfg.before || cfg.after {
			return
		}
		specNode := node{
			text:  text,
			order: cfg.order,
			pend:  cfg.pend,
			focus: cfg.focus,
			index: len(*nodes),
		}
		*nodes = append(*nodes, specNode)
	})
	return *nodes
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
