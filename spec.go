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
	specs, focused, seed := parse(f, opts...)

	t.Logf("Running %d specs.", len(specs))
	if seed != 0 {
		t.Logf("Random seed: %d", seed)
	}

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
			}, func(_ string, f func(), opts ...Option) {
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
