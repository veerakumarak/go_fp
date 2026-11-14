package option

import (
	"fmt"
	"testing"
)

func TestOption(t *testing.T) {
	// 1. Basic creation & unwrapping
	optStr := Some("hello world")
	optEmpty := None[string]()

	if optStr.IsSome() {
		t.Log("the result is", optStr.Unwrap())
	} else {
		panic("result should be some")
	}

	if optEmpty.IsNone() {
		t.Log("the result is none")
	} else {
		panic("result should be none")
	}

	// 2. Pattern Matching via Match()
	optStr.Match(
		func(val string) { fmt.Println("Found: ", val) },
		func() { fmt.Println("Nothing found") },
	)

	// 3. Transform with Map (e.g., string -> int length)
	optLen := Map(Some("hi there"), func(s string) int {
		return len(s)
	})

	if optLen.IsNone() || optLen.Unwrap() != 8 {
		fmt.Println("the result should be of 8 characters, instead found", optLen.Unwrap())
	}

	// 4. Chaining with AndThen
	halfIfEven := func(n int) Option[int] {
		if n%2 == 0 {
			return Some(n / 2)
		}
		return None[int]()
	}

	res1 := AndThen(Some(8), halfIfEven) // Some(4)
	res2 := AndThen(Some(7), halfIfEven) // None

	if res1.IsNone() || res1.Unwrap() != 4 {
		fmt.Println("the result should be Some(4), instead found None")
	}

	if res2.IsSome() {
		fmt.Println("the result should be None, instead found Some")
	}

}
