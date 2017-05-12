package spec

import "testing"

func parse(f func(*testing.T, G, S), opts ...Option) (specs []specInfo, focused bool, seed int64) {
	var groups groupStack
	cfg := options(opts).apply()
	seed = cfg.seed
	groups.primary.order = cfg.order.from(orderSequential)

	// To randomize, store separately until block finishes, then sort
	// TODO: locally sequential blocks
	// TODO: enforce options

	f(nil, func(text string, f func(), opts ...Option) {
		cfg := options(opts).apply()
		groups.push(cfg, text)
		defer groups.pop()
		group := groups.last()
		focused = focused || group.focus && !group.pend
		if groups.orderChanged() {
			prevSpecs := specs
			specs = nil
			defer func() {
				group.order.sort(specs, &seed)
				prevSpecs := append(prevSpecs, specs...)
				specs = prevSpecs
			}()
		}
		f()
	}, func(text string, _ func(), opts ...Option) {
		cfg := options(opts).apply()
		if cfg.before || cfg.after {
			return
		}
		spec := groups.spec(cfg, text)
		focused = focused || spec.focus && !spec.pend
		specs = append(specs, spec)
	})
	return specs, focused, seed
}

type specInfo struct {
	name     []string
	parallel bool
	pend     bool
	focus    bool
	groups   []uint64
	index    uint64
}

type groupInfo struct {
	text       string
	order      order
	pend       bool
	focus      bool
	groupIndex uint64
	specIndex  uint64
}

type groupStack struct {
	groups     []groupInfo
	primary    groupInfo
	groupIndex uint64
	specIndex  uint64
}

func (g *groupStack) push(cfg *config, text string) {
	last := g.last()
	g.groups = append(g.groups, groupInfo{
		text:       text,
		order:      cfg.order.from(last.order),
		pend:       last.pend || cfg.pend,
		focus:      last.focus || cfg.focus,
		groupIndex: g.groupIndex,
		specIndex:  g.specIndex,
	})
	g.groupIndex, g.specIndex = 0, 0
}

func (g *groupStack) pop() {
	l := len(g.groups) - 1
	g.groupIndex = g.groups[l].groupIndex + 1
	g.specIndex = g.groups[l].specIndex
	g.groups = g.groups[:l]
}

func (g *groupStack) last() groupInfo {
	if len(g.groups) == 0 {
		return g.primary
	}
	return g.groups[len(g.groups)-1]
}

func (g *groupStack) orderChanged() bool {
	lastStack := *g
	lastStack.pop()
	return g.last().order != lastStack.last().order
}

func (g *groupStack) spec(cfg *config, text string) specInfo {
	last := g.last()
	spec := specInfo{
		parallel: cfg.order.from(last.order) == orderParallel,
		pend:     cfg.pend || last.pend,
		focus:    cfg.focus || last.focus,
		index:    g.specIndex,
	}
	for _, group := range g.groups {
		spec.name = append(spec.name, group.text)
		spec.groups = append(spec.groups, group.groupIndex)
	}
	spec.name = append(spec.name, text)
	g.specIndex++
	return spec
}
