package spec

import (
	"strings"
	"testing"
)

type G func(string, func(), ...Option)

func (g G) Pend(text string, f func(), _ ...Option) {
	g(text, f, func(c *config) { c.pend = true })
}

func (g G) Focus(text string, f func(), opts ...Option) {
	g(text, f, append(opts, func(c *config) { c.focus = true })...)
}

type S func(string, func(), ...Option)

func (s S) Before(f func()) {
	s("", f, func(c *config) { c.before = true })
}

func (s S) After(f func()) {
	s("", f, func(c *config) { c.after = true })
}

func (s S) Pend(text string, f func(), _ ...Option) {
	s(text, f, func(c *config) { c.pend = true })
}

func (s S) Focus(text string, f func(), opts ...Option) {
	s(text, f, append(opts, func(c *config) { c.focus = true })...)
}

func Run(t *testing.T, f func(*testing.T, G, S), opts ...Option) bool {
	success := true
	specs, focused := parse(f, opts...)

	for _, s := range specs {
		s := s
		name := strings.Join(s.name, "/")
		success = success && t.Run(name, func(t *testing.T) {
			switch {
			case s.pend, focused && !s.focus:
				t.SkipNow()
			case s.parallel:
				t.Parallel()
			}
			var (
				spec          func()
				before, after []func()
			)
			f(t, func(_ string, f func(), _ ...Option) {
				switch {
				case len(s.groups) == 0:
				case s.groups[0] > 0:
					s.groups[0]--
				default:
					s.groups = s.groups[1:]
					f()
				}
			}, func(text string, f func(), opts ...Option) {
				cfg := options(opts).apply()
				switch {
				case cfg.before:
					before = append(before, f)
				case cfg.after:
					after = append([]func(){f}, after...)
				case spec != nil || len(s.groups) > 0:
				case s.index > 0:
					s.index--
				default:
					spec = f
				}
			})

			if spec == nil {
				t.Fatalf("Failed to parse: %s", name)
			}

			run(before...)
			defer run(after...)
			run(spec)
		})
	}

	return success
}

func run(fs ...func()) {
	for _, f := range fs {
		f()
	}
}

type specInfo struct {
	name     []string
	pend     bool
	focus    bool
	parallel bool
	groups   []uint64
	index    uint64
}

func parse(f func(*testing.T, G, S), opts ...Option) (specs []specInfo, focused bool) {
	var (
		global     = options(opts).apply()
		groups     groupStack
		groupIndex uint64
		specIndex  uint64
	)

	f(nil, func(text string, f func(), opts ...Option) {
		cfg := options(opts).apply()

		groups = groups.push(cfg, text, groupIndex, specIndex)
		defer func() { groups, groupIndex, specIndex = groups.pop() }()

		groupIndex, specIndex = 0, 0
		focused = focused || groups.focused()

		f()
	}, func(text string, _ func(), opts ...Option) {
		cfg := options(opts).apply()
		if cfg.before || cfg.after {
			return
		}
		spec := groups.spec(cfg, text, specIndex)

		spec.parallel = spec.parallel || global.parallel
		focused = focused || spec.focus && !spec.pend

		specs = append(specs, spec)
		specIndex++
	})
	return specs, focused
}

type groupInfo struct {
	text       string
	pend       bool
	focus      bool
	parallel   bool
	groupIndex uint64
	specIndex  uint64
}

type groupStack []groupInfo

func (g groupStack) push(cfg *config, text string, groupIndex, specIndex uint64) groupStack {
	group := g.last()
	return append(g, groupInfo{
		text:       text,
		pend:       group.pend || cfg.pend,
		focus:      group.focus || cfg.focus,
		parallel:   group.parallel || cfg.parallel,
		groupIndex: groupIndex,
		specIndex:  specIndex,
	})
}

func (g groupStack) pop() (stack groupStack, groupIndex, specIndex uint64) {
	group := g.last()
	return g[:len(g)-1], group.groupIndex + 1, group.specIndex
}

func (g groupStack) last() groupInfo {
	if len(g) == 0 {
		return groupInfo{}
	}
	return g[len(g)-1]
}

func (g groupStack) focused() bool {
	group := g.last()
	return group.focus && !group.pend
}

func (g groupStack) spec(cfg *config, text string, specIndex uint64) specInfo {
	group := g.last()
	spec := specInfo{
		pend:     cfg.pend || group.pend,
		focus:    cfg.focus || group.focus,
		parallel: cfg.parallel || group.parallel,
		index:    specIndex,
	}
	for _, group := range g {
		spec.name = append(spec.name, group.text)
		spec.groups = append(spec.groups, group.groupIndex)
	}
	spec.name = append(spec.name, text)
	return spec
}

type Option func(*config)

func Parallel() Option {
	return func(c *config) {
		c.parallel = true
	}
}

type config struct {
	pend     bool
	focus    bool
	parallel bool
	before   bool
	after    bool
}

type options []Option

func (o options) apply() *config {
	cfg := &config{}
	for _, opt := range o {
		opt(cfg)
	}
	return cfg
}
