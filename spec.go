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

func (s S) Before(f func()) {
	s("", f, func(c *config) { c.before = true })
}

func (s S) After(f func()) {
	s("", f, func(c *config) { c.after = true })
}

func (s S) Pend(text string, f func(), _ ...Option) {
	s(text, f, func(c *config) { c.pend = true })
}

func Run(t *testing.T, f func(*testing.T, G, S), opts ...Option) bool {
	success := true

	for _, s := range parse(f, opts...) {
		s := s
		name := strings.Join(s.name, "/")
		success = success && t.Run(name, func(t *testing.T) {
			switch {
			case s.pend:
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
	parallel bool
	groups   []uint64
	index    uint64
}

func parse(f func(*testing.T, G, S), opts ...Option) []specInfo {
	type groupInfo struct {
		text       string
		pend       bool
		parallel   bool
		groupIndex uint64
		specIndex  uint64
	}

	var (
		parallel   = options(opts).apply().parallel
		specs      []specInfo
		groups     []groupInfo
		groupIndex uint64 // does this really work?
		specIndex  uint64
	)

	f(nil, func(text string, f func(), opts ...Option) {
		cfg := options(opts).apply()
		groups = append(groups, groupInfo{
			text, cfg.pend,cfg.parallel || parallel,
			groupIndex, specIndex,
		})
		groupIndex, specIndex = 0, 0
		defer func() {
			groupIndex = groups[len(groups)-1].groupIndex + 1
			specIndex = groups[len(groups)-1].specIndex
			groups = groups[:len(groups)-1]
		}()
		f()
	}, func(text string, _ func(), opts ...Option) {
		cfg := options(opts).apply()
		if cfg.before || cfg.after {
			return
		}
		spec := specInfo{pend: cfg.pend, parallel: cfg.parallel || parallel, index: specIndex}
		for _, group := range groups {
			spec.name = append(spec.name, group.text)
			spec.groups = append(spec.groups, group.groupIndex)
			spec.pend = spec.pend || group.pend
			spec.parallel = spec.parallel || group.parallel
		}
		spec.name = append(spec.name, text)
		specs = append(specs, spec)
		specIndex++
	})
	return specs
}

type Option func(*config)

func Parallel() Option {
	return func(c *config) {
		c.parallel = true
	}
}

type config struct {
	pend     bool
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
