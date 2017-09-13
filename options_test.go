package spec_test

import (
	"reflect"
	"testing"

	"github.com/sclevine/spec"
)

func testSpec(name string, it spec.S, s recorder, t *testing.T) {
	if name != "" {
		name += "."
	}
	it.Before(s(t, name+"Before.1"))
	it.Before(s(t, name+"Before.2"))
	it.Before(s(t, name+"Before.3"))
	it.After(s(t, name+"After.1"))
	it.After(s(t, name+"After.2"))
	it.After(s(t, name+"After.3"))
	it(name+"S", s(t, name+"S"))
}

func testSpecs(name string, it spec.S, s recorder, t *testing.T) {
	if name != "" {
		name += "."
	}
	it(name+"S.1", s(t, name+"S.1"))
	it(name+"S.2", s(t, name+"S.2"))
	it(name+"S.3", s(t, name+"S.3"))
}

func testOptionCases(when spec.G, it spec.S, s recorder, t *testing.T) {
	testSpecs("", it, s, t)
	when("G", func() {
		testSpec("G", it, s, t)
	})
	when("G.Sequential", func() {
		testSpecs("G.Sequential", it, s, t)
	}, spec.Sequential())
	when("G.Reverse", func() {
		testSpecs("G.Reverse", it, s, t)
	}, spec.Reverse())
	when("G.Random.Local", func() {
		testSpecs("G.Random.Local", it, s, t)
	}, spec.Random(), spec.Local())
	when("G.Random.Global", func() {
		testSpecs("G.Random.Global", it, s, t)
	}, spec.Random(), spec.Global())
}

func TestSequential(t *testing.T) {
	s, calls := record(t)

	spec.Run(t, "Run", func(t *testing.T, when spec.G, it spec.S) {
		testOptionCases(when, it, s, t)
	}, spec.Sequential(), spec.Seed(2))

	if !reflect.DeepEqual(calls(), []string{
		"Run/S.1->S.1", "Run/S.2->S.2", "Run/S.3->S.3",

		"Run/G/G.S->G.Before.1", "Run/G/G.S->G.Before.2", "Run/G/G.S->G.Before.3",
		"Run/G/G.S->G.S",
		"Run/G/G.S->G.After.1", "Run/G/G.S->G.After.2", "Run/G/G.S->G.After.3",

		"Run/G.Sequential/G.Sequential.S.1->G.Sequential.S.1",
		"Run/G.Sequential/G.Sequential.S.2->G.Sequential.S.2",
		"Run/G.Sequential/G.Sequential.S.3->G.Sequential.S.3",

		"Run/G.Reverse/G.Reverse.S.3->G.Reverse.S.3",
		"Run/G.Reverse/G.Reverse.S.2->G.Reverse.S.2",
		"Run/G.Reverse/G.Reverse.S.1->G.Reverse.S.1",

		"Run/G.Random.Local/G.Random.Local.S.3->G.Random.Local.S.3",
		"Run/G.Random.Local/G.Random.Local.S.1->G.Random.Local.S.1",
		"Run/G.Random.Local/G.Random.Local.S.2->G.Random.Local.S.2",

		"Run/G.Random.Global/G.Random.Global.S.3->G.Random.Global.S.3",
		"Run/G.Random.Global/G.Random.Global.S.1->G.Random.Global.S.1",
		"Run/G.Random.Global/G.Random.Global.S.2->G.Random.Global.S.2",
	}) {
		t.Fatal("Incorrect order:", calls())
	}
}

func TestReverse(t *testing.T) {
	s, calls := record(t)

	spec.Run(t, "Run", func(t *testing.T, when spec.G, it spec.S) {
		testOptionCases(when, it, s, t)
	}, spec.Reverse(), spec.Seed(2))

	if !reflect.DeepEqual(calls(), []string{
		"Run/G.Random.Global/G.Random.Global.S.3->G.Random.Global.S.3",
		"Run/G.Random.Global/G.Random.Global.S.1->G.Random.Global.S.1",
		"Run/G.Random.Global/G.Random.Global.S.2->G.Random.Global.S.2",

		"Run/G.Random.Local/G.Random.Local.S.3->G.Random.Local.S.3",
		"Run/G.Random.Local/G.Random.Local.S.1->G.Random.Local.S.1",
		"Run/G.Random.Local/G.Random.Local.S.2->G.Random.Local.S.2",

		"Run/G.Reverse/G.Reverse.S.3->G.Reverse.S.3",
		"Run/G.Reverse/G.Reverse.S.2->G.Reverse.S.2",
		"Run/G.Reverse/G.Reverse.S.1->G.Reverse.S.1",

		"Run/G.Sequential/G.Sequential.S.1->G.Sequential.S.1",
		"Run/G.Sequential/G.Sequential.S.2->G.Sequential.S.2",
		"Run/G.Sequential/G.Sequential.S.3->G.Sequential.S.3",

		"Run/G/G.S->G.Before.1", "Run/G/G.S->G.Before.2", "Run/G/G.S->G.Before.3",
		"Run/G/G.S->G.S",
		"Run/G/G.S->G.After.1", "Run/G/G.S->G.After.2", "Run/G/G.S->G.After.3",

		"Run/S.3->S.3", "Run/S.2->S.2", "Run/S.1->S.1",
	}) {
		t.Fatal("Incorrect order:", calls())
	}
}
