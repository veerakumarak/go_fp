package failure

import (
	"errors"
	"strings"
	"testing"
)

func TestEmpty(f *testing.T) {
	err := Empty()
	if !err.IsEmpty() {
		f.Errorf("expected Empty() to be empty")
	}
	if err.IsPresent() {
		f.Errorf("expected Empty() not to be present")
	}
	if AsError(err) != nil {
		f.Errorf("expected AsError(Empty()) to return nil")
	}
	if err.Error() != "no failure" {
		f.Errorf("unexpected error message: %s", err.Error())
	}
	if err.String() != "Failure.EMPTY" {
		f.Errorf("unexpected string representation: %s", err.String())
	}
}

func TestNew(f *testing.T) {
	err := New("something went wrong")
	if !err.IsPresent() {
		f.Errorf("expected failure to be present")
	}
	if err.Message() != "something went wrong" {
		f.Errorf("expected message 'something went wrong', got '%s'", err.Message())
	}
	if err.Cause() != nil {
		f.Errorf("expected cause to be nil")
	}
}

func TestWrap(f *testing.T) {
	causeErr := errors.New("root cause")
	wrapped := Wrap("high level failure", causeErr)

	if wrapped.Message() != "high level failure" {
		f.Errorf("unexpected message: %s", wrapped.Message())
	}
	if wrapped.Cause() != causeErr {
		f.Errorf("expected cause to be root cause")
	}
	if wrapped.Unwrap() != causeErr {
		f.Errorf("expected Unwrap() to return root cause")
	}
}

func TestFrom(f *testing.T) {
	// Case 1: nil error
	if !From(nil).IsEmpty() {
		f.Errorf("expected nil error to map to Empty()")
	}

	// Case 2: standard Go error
	stdErr := errors.New("standard error")
	failure := From(stdErr)
	if failure.Message() != "standard error" {
		f.Errorf("expected message 'standard error', got '%s'", failure.Message())
	}

	// Case 3: already a Failure
	original := New("custom failure")
	wrappedAgain := From(original)
	if wrappedAgain != original {
		f.Errorf("expected From on an existing Failure to return it directly")
	}
}

func TestUnwrapFailure(f *testing.T) {
	inner := New("inner failure")
	outer := Wrap("outer failure", inner)

	unwrapped := outer.UnwrapFailure()
	if unwrapped.Message() != "inner failure" {
		f.Errorf("expected recursive unwrap to find 'inner failure', got '%s'", unwrapped.Message())
	}
}

func TestOf(f *testing.T) {
	// Case 1: Success (no panic)
	res := Of(func() {
		// do nothing successfully
	})
	if !res.IsEmpty() {
		f.Errorf("expected successful execution to result in Empty()")
	}

	// Case 2: Panic with string
	resPanicStr := Of(func() {
		panic("something broke")
	})
	if !resPanicStr.IsPresent() || !strings.Contains(resPanicStr.Message(), "something broke") {
		f.Errorf("expected panic string to be caught, got: %s", resPanicStr.Message())
	}

	// Case 3: Panic with error
	resPanicErr := Of(func() {
		panic(errors.New("panic error"))
	})
	if !resPanicErr.IsPresent() || resPanicErr.Cause().Error() != "panic error" {
		f.Errorf("expected panic error to be caught, got: %v", resPanicErr.Cause())
	}
}

func TestInspectorsAndCallbacks(f *testing.T) {
	presentCalled := false
	emptyCalled := false

	failure := New("bad")
	failure.IfPresent(func(err Failure) {
		presentCalled = true
	})

	empty := Empty()
	empty.IfEmpty(func() {
		emptyCalled = true
	})

	if !presentCalled || !emptyCalled {
		f.Errorf("expected inspection callbacks to execute")
	}

	// Test chaining inspect methods
	chained := failure.InspectPresent(func(err Failure) {
		// inspect action
	}).InspectEmpty(func() {
		// should skip
	})

	if chained.IsEmpty() {
		f.Errorf("expected chained failure to remain present")
	}
}

func TestOrThrow(f *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			f.Errorf("expected OrThrow to panic on present failure")
		}
	}()

	failure := New("panic me")
	failure.OrThrow()
}

func TestEqualAndIs(f *testing.T) {
	f1 := New("error")
	f2 := New("error")
	f3 := New("different")

	if !f1.Equal(f2) {
		f.Errorf("expected f1 and f2 to be equal")
	}
	if f1.Equal(f3) {
		f.Errorf("expected f1 and f3 not to be equal")
	}

	// Test errors.Is compatibility
	if !errors.Is(f1, f2) {
		f.Errorf("expected errors.Is to match equal Failures")
	}
}
