package spec

import (
	"testing"
	"time"
)

// G defines a group of specs.
// Unlike other testing libraries, it is re-evaluated for each subspec.
//
// Valid Options:
// Sequential(), Random(), Reverse(), Parallel()
// Local(), Global(), Flat(), Nested()
type G func(text string, f func(), opts ...Option)

// Pend skips all specs in the group.
//
// All Options are ignored.
func (g G) Pend(text string, f func(), _ ...Option) {
	g(text, f, func(c *config) { c.pend = true })
}

// Focus skips all specs except the focused group and other focused specs.
//
// Valid Options:
// Sequential(), Random(), Reverse(), Parallel()
// Local(), Global(), Flat(), Nested()
func (g G) Focus(text string, f func(), opts ...Option) {
	g(text, f, append(opts, func(c *config) { c.focus = true })...)
}

// S defines a spec.
//
// Valid Options: Parallel()
type S func(text string, f func(), opts ...Option)

// Before runs a function before each spec in the group.
func (s S) Before(f func()) {
	s("", f, func(c *config) { c.before = true })
}

// After runs a function after each spec in the group.
func (s S) After(f func()) {
	s("", f, func(c *config) { c.after = true })
}

// Pend skips the spec.
//
// All Options are ignored.
func (s S) Pend(text string, f func(), _ ...Option) {
	s(text, f, func(c *config) { c.pend = true })
}

// Focus skips all specs except the focused spec and other focused specs.
//
// Valid Options: Parallel()
func (s S) Focus(text string, f func(), opts ...Option) {
	s(text, f, append(opts, func(c *config) { c.focus = true })...)
}

// Run defines a suite, which is a top-level group of specs.
// Unlike other testing libraries, it is re-evaluated for each spec.
//
// Valid Options:
// Sequential(), Random(), Reverse(), Parallel()
// Local(), Global(), Flat(), Nested()
func Run(t *testing.T, text string, f func(*testing.T, G, S), opts ...Option) bool {
	cfg := options(opts).apply()
	n := &node{
		text:  []string{text},
		seed:  defaultZero64(cfg.seed, time.Now().Unix()),
		order: cfg.order.or(orderSequential),
		scope: cfg.scope.or(scopeLocal),
		nest:  cfg.nest.or(nestOff),
		pend:  cfg.pend,
		focus: cfg.focus,
	}
	plan := n.parse(f)
	var specs chan Spec
	if cfg.report != nil {
		cfg.report.Start(t, plan)
		specs = make(chan Spec, plan.Total)
		done := make(chan struct{})
		defer func() {
			close(specs)
			<-done
		}()
		go func() {
			cfg.report.Specs(t, specs)
			close(done)
		}()
	}

	return n.run(t, func(t *testing.T, n node) {
		defer func() {
			if specs == nil {
				return
			}
			specs <- Spec{
				Text:     n.text,
				Failed:   t.Failed(),
				Skipped:  t.Skipped(),
				Focused:  n.focus,
				Parallel: n.order == orderParallel,
			}
		}()
		switch {
		case n.pend, plan.HasFocus && !n.focus:
			t.SkipNow()
		case n.order == orderParallel:
			t.Parallel()
		}
		var (
			spec          func()
			before, after []func()
		)
		f(t, func(_ string, f func(), _ ...Option) {
			switch {
			case len(n.loc) == 1, n.loc[0] > 0:
				n.loc[0]--
			default:
				n.loc = n.loc[1:]
				f()
			}
		}, func(_ string, f func(), opts ...Option) {
			cfg := options(opts).apply()
			switch {
			case cfg.before:
				before = append(before, f)
			case cfg.after:
				after = append([]func(){f}, after...)
			case spec != nil:
			case len(n.loc) > 1, n.loc[0] > 0:
				n.loc[0]--
			default:
				spec = f
			}
		})
		if spec == nil {
			t.Fatal("Failed to locate spec.")
		}
		run(before...)
		defer run(after...)
		run(spec)
	})
}

func run(fs ...func()) {
	for _, f := range fs {
		f()
	}
}

// A Plan provides a Reporter with information about a suite.
type Plan struct {
	Text      string
	Total     int
	Pending   int
	Focused   int
	Seed      int64
	HasRandom bool
	HasFocus  bool
};

// A Spec provides a Reporter with information about a spec immediately after
// the spec completes.
type Spec struct {
	Text     []string
	Failed   bool
	Skipped  bool
	Focused  bool
	Parallel bool
}

// A Reporter is provided with information about a suite as it runs.
type Reporter interface {
	Start(*testing.T, Plan)
	Specs(*testing.T, <-chan Spec)
}
