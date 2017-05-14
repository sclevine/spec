package spec

import "time"

// An Option is used to control the behavior of a call to Run, G, or S.
type Option func(*config)

// Parallel indicates that a spec or group of specs should be run in parallel.
// This Option is equivalent to t.Parallel().
// Valid Option for: Run, G, S
func Parallel() Option {
	return func(c *config) {
		c.order = orderParallel
	}
}

// Sequential indicates that a group of specs should be run in order.
// Valid Option for: Run, G
func Sequential() Option {
	return func(c *config) {
		c.order = orderSequential
	}
}

// Random indicates that a group of specs should be run in random order.
// Randomization is per group, such that all groupings are maintained.
// Valid Option for: Run, G
func Random() Option {
	return func(c *config) {
		c.order = orderRandom
	}
}

// Reverse indicates that groups of specs should be run in reverse order.
// Valid Option for: Run, G
func Reverse() Option {
	return func(c *config) {
		c.order = orderReverse
	}
}

// Nest runs each group of specs in a shared subtest.
// This can allows for more control over parallelism.
// Valid Option for: Run, G
func Nest() Option {
	return func(c *config) {
		c.nest = true
	}
}

// Seed specifies the random seed used for all randomized specs.
// The random seed is always displayed before specs are run.
// If not specified, the current time is used.
// Valid Option for: Run
func Seed(s int64) Option {
	return func(c *config) {
		c.seed = s
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

type config struct {
	seed   int64
	order  order
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
