package spec_test

import (
	"fmt"
	"testing"

	"github.com/sclevine/spec"
)

func TestSpec(t *testing.T) {
	spec.Run(t, func(t *testing.T, when spec.G, it spec.S) {
		when("something happens", func() {
			it.Before(func() {
				t.Log("before")
			})

			it.After(func() {
				t.Log("after")
			})

			it("should do something", func() {
				t.Log("first")
				fmt.Println("first")
			})

			when("something nested happens", func() {
				it.Before(func() {
					t.Log("nested before")
				})

				it("should do something nested", func() {
					t.Log("nested")
					fmt.Println("nested")
				})

				it.After(func() {
					t.Log("nested after")
				})
			})

			it("should do something else", func() {
				t.Log("second")
				fmt.Println("second")
			})
		})
	})
}
