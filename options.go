package spec

import (
	"math/rand"
	"time"
)

type Option func(*config)

func Parallel() Option {
	return func(c *config) {
		c.order = orderParallel
	}
}

func Sequential() Option {
	return func(c *config) {
		c.order = orderSequential
	}
}

func Random() Option {
	return func(c *config) {
		c.order = orderRandom
	}
}

func Reverse() Option {
	return func(c *config) {
		c.order = orderReverse
	}
}

func Seed(s int64) Option {
	return func(c *config) {
		c.seed = s
	}
}

func Nest() Option {
	return func(c *config) {
		c.nest = true
	}
}

type order int

const (
	orderInherit order = iota
	orderParallel
	orderSequential
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

func (o order) sort(specs []specInfo, seed *int64) {
	switch o {
	case orderRandom:
		if *seed == 0 {
			*seed = time.Now().Unix()
		}

		r := rand.New(rand.NewSource(*seed))
		for i := len(specs) - 1; i > 0; i-- {
			j := r.Intn(i + 1)
			specs[i], specs[j] = specs[j], specs[i]
		}
	case orderReverse:
		last := len(specs) - 1
		for i := 0; i < len(specs)/2; i++ {
			specs[i], specs[last-i] = specs[last-i], specs[i]
		}
	}
}

type config struct {
	order  order
	seed   int64
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
	return cfg
}
