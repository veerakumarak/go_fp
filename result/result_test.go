package result

import (
	"strconv"
	"testing"

	"github.com/veerakumarak/go_fp/failure"
)

func TestOkAndErr(t *testing.T) {
	okRes := Ok(42)
	if !okRes.IsOk() || okRes.IsErr() {
		t.Errorf("expected Ok result to be valid")
	}
	if okRes.Unwrap() != 42 {
		t.Errorf("expected Unwrap() to return 42, got %d", okRes.Unwrap())
	}

	errRes := Err[int](failure.New("bad state"))
	if errRes.IsOk() || !errRes.IsErr() {
		t.Errorf("expected Err result to be in failure state")
	}
}

func TestErrMsg(t *testing.T) {
	res := ErrMsg[string]("something went wrong")
	if !res.IsErr() {
		t.Errorf("expected ErrMsg to be an error result")
	}
	if res.Failure().Message() != "something went wrong" {
		t.Errorf("unexpected failure message: %s", res.Failure().Message())
	}
}

func TestFromGo(t *testing.T) {
	// Success case
	okRes := FromGo(strconv.Atoi("100"))
	if !okRes.IsOk() || okRes.Unwrap() != 100 {
		t.Errorf("expected successful FromGo result")
	}

	// Error case
	errRes := FromGo(strconv.Atoi("invalid"))
	if !errRes.IsErr() {
		t.Errorf("expected failed FromGo result")
	}
}

func TestOf(t *testing.T) {
	// Success
	res := Of(func() (int, error) {
		return 50, nil
	})
	if res.Unwrap() != 50 {
		t.Errorf("expected 50, got %d", res.Unwrap())
	}

	// Panic Recovery
	panicRes := Of(func() (int, error) {
		panic("kaboom")
	})
	if !panicRes.IsErr() {
		t.Errorf("expected panic to be caught as a failure")
	}
}

func TestUnwrapOrAndOrElse(t *testing.T) {
	okRes := Ok(10)
	errRes := ErrMsg[int]("fail")

	if okRes.UnwrapOr(99) != 10 {
		t.Errorf("expected 10 for Ok result UnwrapOr")
	}
	if errRes.UnwrapOr(99) != 99 {
		t.Errorf("expected fallback 99 for Err result UnwrapOr")
	}

	// UnwrapOrElse with supplier
	fallback := errRes.UnwrapOrElse(func() int {
		return 500
	})
	if fallback != 500 {
		t.Errorf("expected supplier fallback 500, got %d", fallback)
	}
}

func TestExpectAndUnwrapPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected Expect to panic on error result")
		}
	}()

	errRes := ErrMsg[int]("critical error")
	errRes.Expect("failed to load user configuration")
}

func TestMapAndFlatMap(t *testing.T) {
	// Map success
	res := Ok(5)
	mapped := Map(res, func(n int) string {
		return strconv.Itoa(n * 2)
	})
	if mapped.Unwrap() != "10" {
		t.Errorf("expected '10', got '%s'", mapped.Unwrap())
	}

	// FlatMap success
	flatMapped := FlatMap(res, func(n int) Result[int] {
		return Ok(n + 5)
	})
	if flatMapped.Unwrap() != 10 {
		t.Errorf("expected 10, got %d", flatMapped.Unwrap())
	}

	// Map error propagation
	errRes := ErrMsg[int]("error")
	mappedErr := Map(errRes, func(n int) int { return n * 2 })
	if !mappedErr.IsErr() {
		t.Errorf("expected error to propagate through Map")
	}
}

func TestMatch(t *testing.T) {
	okCalled := false
	errCalled := false

	Ok("hello").Match(
		func(val string) { okCalled = true },
		func(f failure.Failure) { errCalled = true },
	)

	if !okCalled || errCalled {
		t.Errorf("expected onOk callback to execute")
	}
}

func TestEqual(t *testing.T) {
	r1 := Ok(42)
	r2 := Ok(42)
	r3 := Ok(100)

	if !r1.Equal(r2) {
		t.Errorf("expected r1 and r2 to be equal")
	}
	if r1.Equal(r3) {
		t.Errorf("expected r1 and r3 not to be equal")
	}
}
