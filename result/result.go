package result

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/veerakumarak/go_fp/failure"
)

type Result[T any] struct {
	value   T
	failure failure.Failure
}

func Ok[T any](value T) Result[T] {
	return Result[T]{
		value:   value,
		failure: failure.Empty(),
	}
}

// Err creates a failed result wrapper.
func Err[T any](f failure.Failure) Result[T] {
	return Result[T]{failure: f}
}

// ErrMsg is a convenience constructor to create a failed Result directly from a message.
func ErrMsg[T any](msg string) Result[T] {
	return Err[T](failure.New(msg))
}

// FromGo bridges standard Go (T, error) functions directly into a Result[T].
func FromGo[T any](val T, err error) Result[T] {
	if err != nil {
		// If it's already a Failure, preserve it
		if f, ok := errors.AsType[failure.Failure](err); ok {
			return Err[T](f)
		}
		return Err[T](failure.Wrap(err.Error(), err))
	}
	return Ok(val)
}

// Of executes a supplier function returning (T, error), safely converting
// any returned error into a Failure and recovering from unexpected panics.
func Of[T any](supplier func() (T, error)) (res Result[T]) {
	if supplier == nil {
		panic("supplier cannot be nil")
	}

	defer func() {
		if r := recover(); r != nil {
			var err error
			switch val := r.(type) {
			case error:
				err = val
			case string:
				err = errors.New(val)
			default:
				err = fmt.Errorf("%v", r)
			}
			res = Err[T](failure.Wrap("An unexpected panic occurred during execution.", err))
		}
	}()

	val, err := supplier()
	if err != nil {
		return Err[T](failure.From(err))
	}

	return Ok(val)
}

func (r Result[T]) IsOk() bool {
	return r.failure.IsEmpty()
}

func (r Result[T]) IsErr() bool {
	return !r.IsOk()
}

// Expect returns the inner value if Ok, or panics with a custom message
// combined with the underlying Failure if Err.
func (r Result[T]) Expect(msg string) T {
	if r.IsErr() {
		// Wrap or format the panic message with context
		panic(failure.Wrap(msg, r.failure))
		//panic(failure.Wrap(message+": "+r.failure.Message(), r.failure))
	}
	return r.value
}

// Unwrap returns the inner value if Ok, or panics with the Failure if Err.
// (Functional equivalent to Get())
func (r Result[T]) Unwrap() T {
	if r.IsErr() {
		panic(r.failure)
	}
	return r.value
}

func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.IsOk() {
		return r.value
	}

	return defaultValue
}

func (r Result[T]) UnwrapOrElse(supplier func() T) T {
	if r.IsOk() {
		return r.value
	}

	if supplier == nil {
		panic("supplier cannot be nil")
	}

	return supplier()
}

func (r Result[T]) Failure() failure.Failure {
	if r.IsOk() {
		panic("cannot get failure from a successful Result")
	}

	return r.failure
}

func (r Result[T]) IfOk(action func(T)) {
	if action == nil {
		panic("action cannot be nil")
	}

	if r.IsOk() {
		action(r.value)
	}
}

func (r Result[T]) IfFailure(action func(failure.Failure)) {
	if action == nil {
		panic("action cannot be nil")
	}

	if r.IsErr() {
		action(r.failure)
	}
}

func (r Result[T]) InspectOk(action func(T)) Result[T] {
	if action == nil {
		panic("action cannot be nil")
	}

	if r.IsOk() {
		action(r.value)
	}

	return r
}

func (r Result[T]) InspectFailure(action func(failure.Failure)) Result[T] {
	if action == nil {
		panic("action cannot be nil")
	}

	if r.IsErr() {
		action(r.failure)
	}

	return r
}

func (r Result[T]) String() string {
	if r.IsOk() {
		return fmt.Sprintf("Result.Ok(%v)", r.value)
	}

	return fmt.Sprintf("Result.Failure(%v)", r.failure)
}

func (r Result[T]) Equal(other Result[T]) bool {
	if r.IsOk() && other.IsOk() {
		return reflect.DeepEqual(r.value, other.value)
	}

	if r.IsErr() && other.IsErr() {
		return r.failure.Equal(other.failure)
	}

	return false
}

// Map transforms the success value using the given function, propagating any Failure untouched.
func Map[T any, R any](r Result[T], fn func(T) R) Result[R] {
	if r.IsErr() {
		return Err[R](r.failure)
	}
	if fn == nil {
		panic("mapping function cannot be nil")
	}
	return Ok(fn(r.value))
}

// FlatMap chains an operation that returns another Result[R], propagating any Failure untouched.
func FlatMap[T any, R any](r Result[T], fn func(T) Result[R]) Result[R] {
	if r.IsErr() {
		return Err[R](r.failure)
	}
	if fn == nil {
		panic("flatmapping function cannot be nil")
	}
	return fn(r.value)
}

// Match executes the onOk callback if the Result is successful,
// or the onFailure callback if it contains a Failure.
func (r Result[T]) Match(onOk func(T), onFailure func(failure.Failure)) {
	if r.IsOk() {
		if onOk == nil {
			panic("onOk callback cannot be nil")
		}
		onOk(r.value)
	} else {
		if onFailure == nil {
			panic("onFailure callback cannot be nil")
		}
		onFailure(r.failure)
	}
}
