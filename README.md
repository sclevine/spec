# spec

Spec is a simple and robust BDD test organizer for Go. Spec is a minimal, additive
extension of the standard library `testing` package. Spec can be considered a
lightweight wrapper for Go 1.7+ [subtests](https://blog.golang.org/subtests).

Spec differs from other BDD libraries for Go in that it:
- Does not re-implement or replace any functionality of the `testing` package
- Does not provide assertions
- Does not encourage the use of dot-imports
- Does not re-use any closures between specs (to avoid test pollution)
- Is implemented without interface types, reflection, global state, locks, or goroutines

Features:
- Clean, simple, straightforward syntax
- Wraps the Go 1.7+ [subtest](https://blog.golang.org/subtests) functionality of the `testing` package
- Supports focusing and pending tests
- Supports random test order
- Supports parallel testing