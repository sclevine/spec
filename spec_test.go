package spec_test

import (
	"testing"
	"reflect"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func record(t *testing.T) (s func(*testing.T) func(), calls func() []string) {
	var names []string
	return func(ts *testing.T) func() {
		return func() {
			if ts == nil {
				t.Fatal("Spec running during parse phase for:", t.Name())
			}
			names = append(names, ts.Name())
		}
	}, func() []string {
		return names
	}
}

func TestPend(t *testing.T) {
	s, calls := record(t)

	spec.Pend(t, "Pend", func(t *testing.T, when spec.G, it spec.S) {
		it("S", s(t))
		it.Pend("S.Pend", s(t))
		it.Focus("S.Focus", s(t))

		when("G", func() {
			it("S", s(t))
		})
		when.Pend("G.Pend", func() {
			it("S", s(t))
		})
		when.Focus("G.Focus", func() {
			it("S", s(t))
		})
	})
	
	if len(calls()) != 0 {
		t.Fatal("Failed to pend:", calls())
	}
}

func TestGPend(t *testing.T) {
	s, calls := record(t)
	
	spec.Run(t, "Run", func(t *testing.T, when spec.G, it spec.S) {
		when.Pend("G.Pend", func() {
			it("S", s(t))
			it.Pend("S.Pend", s(t))
			it.Focus("S.Focus", s(t))

			when("G", func() {
				it("S", s(t))
			})
			when.Pend("G.Pend", func() {
				it("S", s(t))
			})
			when.Focus("G.Focus", func() {
				it("S", s(t))
			})
		})
	})
	
	if len(calls()) != 0 {
		t.Fatal("Failed to pend:", calls())
	}
}

func TestSPend(t *testing.T) {
	s, calls := record(t)
	
	spec.Run(t, "Run", func(t *testing.T, when spec.G, it spec.S) {
		it.Pend("S", s(t))
	})
	
	if len(calls()) != 0 {
		t.Fatal("Failed to pend:", calls())
	}
}

func testCases(when spec.G, it spec.S, f func()) {
	it.Before(f)
	it.After(f)

	it("S", f)
	it.Pend("S.Pend", f)
	it.Focus("S.Focus", f)

	when("G", func() {
		it.Before(f)
		it.After(f)
		it("S", f)
	})
	when.Pend("G.Pend", func() {
		it.Before(f)
		it.After(f)
		it("S", f)
	})
	when.Focus("G.Focus", func() {
		it.Before(f)
		it.After(f)
		it("S", f)
	})
}

func TestFocus(t *testing.T) {
	s, calls := record(t)
	
	spec.Focus(t, "Focus", func(t *testing.T, when spec.G, it spec.S) {
		it("S", s(t))
		it.Pend("S.Pend", s(t))
		it.Focus("S.Focus", s(t))

		when("G", func() {
			it("S", s(t))
		})
		when.Pend("G.Pend", func() {
			it("S", s(t))
		})
		when.Focus("G.Focus", func() {
			it("S", s(t))
		})
	})
	
	if !reflect.DeepEqual(calls(), []string{
		"Focus/S", "Focus/S.Focus",
		"Focus/G/S", "Focus/G.Focus/S",
	}) {
		t.Fatal("Incorrect focus:", calls())
	}
}

func TestGFocus(t *testing.T) {
	s, calls := record(t)
	
	spec.Run(t, "Run", func(t *testing.T, when spec.G, it spec.S) {
		when.Focus("G.Focus", func() {
			it("S", s(t))
			it.Pend("S.Pend", s(t))
			it.Focus("S.Focus", s(t))

			when("G", func() {
				it("S", s(t))
			})
			when.Pend("G.Pend", func() {
				it("S", s(t))
			})
			when.Focus("G.Focus", func() {
				it("S", s(t))
			})
		})
	})
	
	if len(calls()) > 0 {
		t.Fatal("Failed to Focus:", calls())
	}
}

func TestSFocus(t *testing.T) {
	s, calls := record(t)
	
	spec.Run(t, "Run", func(t *testing.T, when spec.G, it spec.S) {
		it.Focus("S", s(t))
	})
	
	if len(calls()) > 0 {
		t.Fatal("Failed to Focus:", calls())
	}
}

func TestSpec(t *testing.T) {
	spec.Run(t, "spec", func(t *testing.T, when spec.G, it spec.S) {
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

			when("some things happen in reverse and in nested subtests", func() {
				it.Before(func() {
					t.Log("before reverse")
				})

				it("should do another thing second", func() {
					t.Log("second reverse")
				})

				it("should do one thing first", func() {
					t.Log("first reverse")
				})
			}, spec.Reverse(), spec.Nested())

			when("some things happen in globally random order", func() {
				it.Before(func() {
					t.Log("before global")
				})

				when("grouped first", func() {
					it.Before(func() {
						t.Log("before group one global")
					})

					it("should do one thing in group one randomly", func() {
						t.Log("group one, spec one, global random")
					})

					it("should do another thing in group one randomly", func() {
						t.Log("group one, spec two, global random")
					})
				})

				when("grouped second", func() {
					it.Before(func() {
						t.Log("before group two global")
					})

					it("should do one thing in group two randomly", func() {
						t.Log("group two, spec one, global random")
					})

					it("should do another thing in group two randomly", func() {
						t.Log("group two, spec two, global random")
					})
				}, spec.Local())

				it("should do one thing ungrouped", func() {
					t.Log("ungrouped global random")
				})
			}, spec.Random(), spec.Global())

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
	}, spec.Report(report.Terminal{}))
}
