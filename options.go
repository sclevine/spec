package spec

import "time"

// An Option is used to control the behavior of a call to Run, G, or S.
type Option func(*config)

// Nest runs each group of spec subtests in a parent subtest.
// This allows for more control over parallelism.
//
// Valid Option for: Run, G
func Nest() Option {
	return func(c *config) {
		c.nest = true
	}
}

// Seed specifies the random seed used for any randomized specs in a Run block.
// The random seed is always displayed before specs are run.
// If not specified, the current time is used.
//
// Valid Option for: Run
func Seed(s int64) Option {
	return func(c *config) {
		c.seed = s
	}
}

// Sequential indicates that a group of specs should be run in order.
// This is the default behavior.
//
// Valid Option for: Run, G
func Sequential() Option {
	return func(c *config) {
		c.order = orderSequential
	}
}

// Random indicates that a group of specs should be run in random order.
// Randomization is per group, such that all groupings are maintained.
//
// Valid Option for: Run, G
func Random() Option {
	return func(c *config) {
		c.order = orderRandom
	}
}

// Reverse indicates that a group of specs should be run in reverse order.
//
// Valid Option for: Run, G
func Reverse() Option {
	return func(c *config) {
		c.order = orderReverse
	}
}

// Parallel indicates that a spec or group of specs should be run in parallel.
// This Option is equivalent to t.Parallel().
//
// Valid Option for: Run, G, S
func Parallel() Option {
	return func(c *config) {
		c.order = orderParallel
	}
}

// Local indicates that the test order applies to each subgroup individually.
// For example, a group with Random() and Local() will run all subgroups and
// specs in random order, and each subgroup will be randomized, but specs in
// different subgroups will not be interleaved.
// This is the default behavior.
//
// Valid Option for: Run, G
func Local() Option {
	return func(c *config) {
		c.scope = scopeLocal
	}
};

// Global indicates that test order applies globally to all descendant specs.
// For example, a group with Random() and Global() will run all descendant
// specs in random order, regardless of subgroup. Specs in different subgroups
// may be interleaved.
//
// Valid Option for: Run, G
func Global() Option {
	return func(c *config) {
		c.scope = scopeGlobal
	}
}

type order int

const (
	orderInherit order = iota
	orderSequential
	orderParallel
	orderRandom
	orderReverse
)

func (o order) from(last order) order {
	switch o {
	case orderInherit:
		return last
	default:
		return o
	}
}

type scope int

const (
	scopeInherit scope = iota
	scopeLocal
	scopeGlobal
)

func (s scope) from(last scope) scope {
	switch s {
	case scopeInherit:
		return last
	default:
		return s
	}
}

type config struct {
	seed   int64
	order  order
	scope  scope
	nest   bool
	pend   bool
	focus  bool
	before bool
	after  bool
}

type options []Option

func (o options) apply() *config {
	cfg := &config{}
	for _, opt := range o {
		opt(cfg)
	}
	if cfg.seed == 0 {
		cfg.seed = time.Now().Unix()
	}
	return cfg
}
