package spec

import "testing"

// G is a grouping of specs.
// Unlike other testing libraries, it is re-evaluated for each spec.
// Valid Options: Parallel(), Sequential(), Random(), Reverse(), Nest()
type G func(text string, f func(), opts ...Option)

// Pend skips all specs in a grouping of specs.
// All Options are ignored.
func (g G) Pend(text string, f func(), _ ...Option) {
	g(text, f, func(c *config) { c.pend = true })
}

// Focus skips all specs except the focused grouping and other focused specs.
// Valid Options: Parallel(), Sequential(), Random(), Reverse(), Nest()
func (g G) Focus(text string, f func(), opts ...Option) {
	g(text, f, append(opts, func(c *config) { c.focus = true })...)
}

// S is a spec.
// Valid Options: Parallel()
type S func(text string, f func(), opts ...Option)

// Before runs before each spec in the group.
func (s S) Before(f func()) {
	s("", f, func(c *config) { c.before = true })
}

// After runs after each spec in the group.
func (s S) After(f func()) {
	s("", f, func(c *config) { c.after = true })
}

// Pend skips the spec.
// All Options are ignored.
func (s S) Pend(text string, f func(), _ ...Option) {
	s(text, f, func(c *config) { c.pend = true })
}

// Focus skips all specs except the focused spec and other focused specs.
// Valid Options: Parallel()
func (s S) Focus(text string, f func(), opts ...Option) {
	s(text, f, append(opts, func(c *config) { c.focus = true })...)
}

// Run is a top-level grouping of specs.
// Unlike other testing libraries, it is re-evaluated for each spec.
// Valid Options: Parallel(), Sequential(), Random(), Reverse(), Nest(), Seed()
func Run(t *testing.T, f func(*testing.T, G, S), opts ...Option) bool {
	cfg := options(opts).apply()
	n := &node{
		seed:  cfg.seed,
		order: cfg.order.from(orderSequential),
		pend:  cfg.pend,
		focus: cfg.focus,
	}
	sum := n.parse(f)
	t.Logf("Total: %d | Focused: %d | Pending: %d", sum.total, sum.focused, sum.pending)
	if sum.random {
		t.Logf("Random seed: %d", cfg.seed)
	}
	if sum.focus {
		t.Log("Focus is active.")
	}

	return n.nodes.run(t, func(t *testing.T, n node) {
		switch {
		case n.pend, sum.focus && !n.focus:
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
			case len(n.loc) < 2, n.loc[0] > 0:
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
