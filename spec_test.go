package spec_test

import (
	"testing"

	"github.com/sclevine/spec"
)

func TestSpec(t *testing.T) {
	spec.Run(t, func(t *testing.T, when spec.G, it spec.S) {
		when("something happens", func() {
			var someStr string

			it.Before(func() {
				t.Log("before")
				if someStr == "some-value" {
					t.Fatal("test pollution")
				}
				someStr = "some-value"
			})

			it.After(func() {
				t.Log("after")
			})

			it("should do something", func() {
				t.Log("first")
			})

			when("something else also happens", func() {
				it.Before(func() {
					t.Log("nested before")
				})

				it("should do something nested", func() {
					t.Log("second")
				})

				it.After(func() {
					t.Log("nested after")
				})
			})

			when("some things happen in parallel at the end", func() {
				it.After(func() {
					t.Log("lone after")
				})

				it("should do one thing in parallel", func() {
					t.Log("first parallel")
				})

				it("should do another thing in parallel", func() {
					t.Log("second parallel")
				})
			}, spec.Parallel())

			when("some things happen randomly", func() {
				it.Before(func() {
					t.Log("before random")
				})

				it("should do one thing randomly", func() {
					t.Log("first random")
				})

				it("should do another thing randomly", func() {
					t.Log("second random")
				})
			}, spec.Random())

			when("some things happen in reverse", func() {
				it.Before(func() {
					t.Log("before reverse")
				})

				it("should do another thing second", func() {
					t.Log("second reverse")
				})

				it("should do one thing first", func() {
					t.Log("first reverse")
				})
			}, spec.Reverse())

			it("should do something else", func() {
				t.Log("third")
			})

			it.Pend("should not do this", func() {
				t.Log("forth")
			})

			when.Pend("nothing important happens", func() {
				it.Focus("should not really focus on this", func() {
					t.Log("fifth")
				})
			})
		})
	})
}
