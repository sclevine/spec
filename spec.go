package spec

import (
	"strings"
	"testing"
)

type G func(string, func(), ...Option)

func (g G) Pend(text string, f func(), _ ...Option) {
	g(text, f, func(c *config) { c.pend = true })
}

type S func(string, func(), ...Option)

func (s S) Before(f func(), opts ...Option) {
	s("", f, append(opts, func(c *config) { c.before = true })...)
}

func (s S) After(f func(), opts ...Option) {
	s("", f, append(opts, func(c *config) { c.after = true })...)
}

func (s S) Pend(text string, f func(), _ ...Option) {
	s(text, f, func(c *config) { c.pend = true })
}

func Run(t *testing.T, f func(*testing.T, G, S)) bool {
	success := true
	specs := parse(f)

	for i := range specs {
		s := specs[i]
		if ok := t.Run(strings.Join(s.name, "/"), func(t *testing.T) {
			t.Parallel()
			var (
				spec          func()
				before, after []func()
			)
			f(t, func(_ string, f func(), _ ...Option) {
				scanGroups := len(s.groups) > 0
				switch {
				case scanGroups && s.groups[0] == 0:
					s.groups = s.groups[1:]
					f()
				case scanGroups:
					s.groups[0]--
				}
			}, func(text string, f func(), opts ...Option) {
				cfg := options(opts).apply()
				scanGroup := spec == nil && len(s.groups) == 0
				switch {
				case cfg.before:
					before = append(before, f)
				case cfg.after:
					after = append([]func(){f}, after...)
				case scanGroup && s.index == 0:
					spec = f
				case scanGroup:
					s.index--
				}
			})

			run(before...)
			defer run(after...)
			run(spec)
		}); !ok {
			success = false
		}
	}

	return success
}

func run(fs ...func()) {
	for _, f := range fs {
		f()
	}
}

type specInfo struct {
	name   []string
	pend   bool
	groups []uint64
	index  uint64
}

type groupInfo struct {
	text       string
	pend       bool
	groupIndex uint64
	specIndex  uint64
}

func parse(f func(*testing.T, G, S)) []specInfo {
	var (
		specs          []specInfo
		groups         []groupInfo
		nextGroupIndex uint64 // does this really work?
		nextSpecIndex  uint64
	)

	f(nil, func(text string, f func(), opts ...Option) {
		pend := options(opts).apply().pend
		groups = append(groups, groupInfo{text, pend, nextGroupIndex, nextSpecIndex})
		nextGroupIndex = 0
		nextSpecIndex = 0
		defer func() {
			nextGroupIndex = groups[len(groups)-1].groupIndex + 1
			nextSpecIndex = groups[len(groups)-1].specIndex
			groups = groups[:len(groups)-1]
		}()
		f()
	}, func(text string, _ func(), opts ...Option) {
		cfg := options(opts).apply()
		if cfg.before || cfg.after {
			return
		}
		spec := specInfo{pend: cfg.pend, index: nextSpecIndex}
		for _, group := range groups {
			spec.name = append(spec.name, group.text)
			spec.groups = append(spec.groups, group.groupIndex)
			spec.pend = spec.pend || group.pend
		}
		spec.name = append(spec.name, text)
		specs = append(specs, spec)
		nextSpecIndex++
	})
	return specs
}

type Option func(*config)

type config struct {
	pend   bool
	before bool
	after  bool
}

type options []Option

func (o options) apply() *config {
	cfg := &config{}
	for _, opt := range o {
		opt(cfg)
	}
	return cfg
}
