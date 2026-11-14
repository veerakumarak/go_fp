# Go FP (Functional Programming Toolkit)

`go_fp` is a lightweight, zero-dependency, idiomatic functional programming library for Go (`1.26+`). It brings expressive functional ergonomics—inspired by Rust, Scala, and Java's `Optional`/`Result` patterns—into Go while maintaining 100% compatibility with Go's standard library (`errors.Is`, `errors.As`, and `(T, error)` tuple returns).

---

## Why `go_fp`?

While Go's explicit `if err != nil` handling is simple, deeply nested validation pipelines, multi-step transformations, and nullable pointer dereferences can quickly clutter business logic. `go_fp` solves this without fighting the language:

- **Railway-Oriented Programming:** Chain operations cleanly with `Map` and `FlatMap`; errors automatically short-circuit down the pipeline without repetitive boilerplate.
- **Zero-Panic Safety:** Wrap risky operations or third-party functions with `Of` to automatically catch unexpected panics and convert them into structured `Failure` values.
- **Standard Library Interop:** `Failure` implements Go's native `error`, `Unwrap()`, and `Is()` interfaces, allowing seamless integration with existing Go packages.
- **Explicit Nullability:** Eliminate `nil` pointer dereference bugs by modeling optional values explicitly with `Option[T]`.

---

## Installation

```bash
go get github.com/veerakumarak/go_fp
```

---

## Package Overview

| Package | Primary Type | Purpose |
| :--- | :--- | :--- |
| `failure` | `Failure` | Rich, chainable domain error abstraction compatible with Go's `errors` package. |
| `result` | `Result[T]` | Represents either a success value (`Ok`) or an error state (`Failure`). |
| `option` | `Option[T]` | Represents an optional value (`Some` or `None`) without relying on `nil` pointers. |
| `pair` | `Pair[A, B]` | Immutable, generic 2-tuple container for grouping related values. |

---

## Detailed Guides & Use Cases

### 1. Failure Handling (`failure`)

`Failure` is a structured error type holding a human-readable message and an optional underlying `cause`. It avoids `nil` error ambiguity by providing a sentinel `Empty()` state.

#### Creating & Wrapping Errors
```go
package main

import (
	"errors"
	"fmt"
	"github.com/veerakumarak/go_fp/failure"
)

var ErrNotFound = errors.New("record not found")

func fetchUser(id string) failure.Failure {
	if id == "" {
		// Create a standalone Failure
		return failure.New("user ID cannot be empty")
	}

	// Wrap an existing standard Go error with domain context
	return failure.Wrap("failed to load user profile", ErrNotFound)
}
```

#### Standard Library Interoperability (`errors.Is` / `errors.As`)
Because `Failure` implements `Unwrap() error` and `Is(error) bool`, it works seamlessly with Go's standard error inspection tools:

```go
f := fetchUser("123")

// Convert Failure back to a standard Go error (returns nil if f.IsEmpty())
var err error = failure.AsError(f)

if errors.Is(err, ErrNotFound) {
	fmt.Println("Handle 404 Not Found gracefully")
}

// Convert any standard Go error into a Failure safely
converted := failure.FromError(err)
fmt.Println(converted.Message()) // "failed to load user profile"
```

---

### 2. Result Monad (`result`)

`Result[T]` encapsulates the outcome of an operation that may succeed with a value of type `T` or fail with a `Failure`.

#### Bridging Standard Go Functions (`FromGo` & `Of`)
You don't need to rewrite your existing Go codebase to use `Result[T]`:

```go
package main

import (
	"strconv"
	"github.com/veerakumarak/go_fp/result"
)

// 1. Bridge standard (T, error) tuple returns directly
portRes := result.FromGo(strconv.Atoi("8080"))

// 2. Execute risky functions with automatic panic recovery
safeRes := result.Of(func() (string, error) {
	// Even if a panic occurs here, Of() catches it and returns a Failure Result
	return riskyExternalCall()
})
```

#### Real-World Use Case: Validation & Transformation Pipeline
Chain parsing, validation, and formatting without a single `if err != nil` block:

```go
package main

import (
	"fmt"
	"strconv"
	"strings"
	"github.com/veerakumarak/go_fp/failure"
	"github.com/veerakumarak/go_fp/result"
)

type DiscountCode struct {
	Code       string
	Percentage int
}

func parseDiscount(rawCode string, rawPercent string) result.Result[DiscountCode] {
	// Step 1: Parse integer from string
	percentRes := result.FromGo(strconv.Atoi(rawPercent))

	// Step 2: Validate business rules using FlatMap
	validPercent := result.FlatMap(percentRes, func(p int) result.Result[int] {
		if p <= 0 || p > 100 {
			return result.ErrMsg[int]("discount percentage must be between 1 and 100")
		}
		return result.Ok(p)
	})

	// Step 3: Transform into domain struct using Map
	return result.Map(validPercent, func(p int) DiscountCode {
		return DiscountCode{
			Code:       strings.ToUpper(strings.TrimSpace(rawCode)),
			Percentage: p,
		}
	})
}

func main() {
	res := parseDiscount("  save20 ", "20")

	// Pattern match on the final outcome
	res.Match(
		func(dc DiscountCode) {
			fmt.Printf("Applied %s for %d%% off!\n", dc.Code, dc.Percentage)
		},
		func(f failure.Failure) {
			fmt.Printf("Invalid discount: %s\n", f.Message())
		},
	)
}
```

#### Unwrapping & Side Effects
`Result[T]` provides ergonomic extraction and inspection methods:

```go
res := result.Ok(100)

// Safe extraction with fallbacks
val1 := res.UnwrapOr(0)                          // 100
val2 := res.UnwrapOrElse(func() int { return 0 }) // Lazy evaluation on failure

// Panic extraction (useful in tests or startup initialization)
val3 := res.Unwrap()
val4 := res.Expect("critical config value missing")

// Fluent side-effect inspection (great for logging/metrics)
res.
	InspectOk(func(v int) { fmt.Println("Audit log success:", v) }).
	InspectFailure(func(f failure.Failure) { fmt.Println("Audit log error:", f) })
```

---

### 3. Option Monad (`option`)

`Option[T]` models the presence (`Some[T]`) or absence (`None[T]`) of a value, making optional fields, map lookups, and configuration overrides explicit and safe.

#### Real-World Use Case: Configuration Lookup & Filtering
```go
package main

import (
	"fmt"
	"os"
	"strings"
	"github.com/veerakumarak/go_fp/option"
)

func getEnvOption(key string) option.Option[string] {
	val, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(val) == "" {
		return option.None[string]()
	}
	return option.Some(val)
}

func main() {
	// Safely look up, trim, and fallback without nil checks
	apiHost := option.Map(getEnvOption("API_HOST"), strings.TrimSpace).
		UnwrapOr("https://api.example.com")

	fmt.Println("Connecting to:", apiHost)
}
```

---

### 4. Pair (`pair`)

`Pair[A, B]` is a lightweight, type-safe tuple for returning or storing two associated values without defining one-off structs.

#### Real-World Use Case: Returning Key-Value Metadata or Pagination Bounds
```go
package main

import (
	"fmt"
	"github.com/veerakumarak/go_fp/pair"
)

func calculatePagination(page, pageSize int) pair.Pair[int, int] {
	offset := (page - 1) * pageSize
	limit := pageSize
	return pair.New(offset, limit)
}

func main() {
	bounds := calculatePagination(3, 25)
	fmt.Printf("SQL Query -> OFFSET %d LIMIT %d\n", bounds.First(), bounds.Second())

	// Equality comparison for comparable types
	other := pair.New(50, 25)
	if pair.Equal(bounds, other) {
		fmt.Println("Pagination bounds match!")
	}
}
```

---

## Contributing

Contributions, bug reports, and feature requests are welcome! Feel free to open an issue or submit a pull request on GitHub.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.