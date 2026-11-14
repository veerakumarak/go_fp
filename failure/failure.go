package failure

import (
	"errors"
	"fmt"
	"strings"
)

type Failure struct {
	message string
	cause   error
}

var emptyFailure = Failure{
	message: "no failure",
	cause:   nil,
}

func Empty() Failure {
	return emptyFailure
}

// New creates a failure with just a message.
func New(message string) Failure {
	return Failure{message: message}
}

// Wrap creates a failure with a custom message and an underlying cause.
func Wrap(message string, cause error) Failure {
	finalMessage := strings.TrimSpace(message)

	if finalMessage == "" {
		if cause != nil && strings.TrimSpace(cause.Error()) != "" {
			finalMessage = cause.Error()
		} else {
			finalMessage = "An unexpected error occurred."
		}
	}

	return Failure{message: finalMessage, cause: cause}
}

// From converts a standard Go error directly into a Failure.
func From(err error) Failure {
	if err == nil {
		return emptyFailure
	}

	// If it's already our Failure type, return it directly
	if f, ok := errors.AsType[Failure](err); ok {
		return f
	}

	msg := err.Error()
	if strings.TrimSpace(msg) == "" {
		msg = "unknown failure occurred"
	}

	return Failure{message: msg, cause: err}
}

// AsError converts a Failure into a standard Go error,
// safely returning nil if the failure is empty.
func AsError(f Failure) error {
	if f.IsEmpty() {
		return nil
	}
	return f
}

// Error implements the standard Go error interface.
func (f Failure) Error() string {
	return f.Message()
}

func (f Failure) Message() string {
	if f == emptyFailure {
		return f.message
	}

	ownMessage := strings.TrimSpace(f.message)
	if ownMessage != "" {
		return f.message
	}

	if f.cause != nil {
		causeMessage := strings.TrimSpace(f.cause.Error())
		if causeMessage != "" {
			return f.cause.Error()
		}
	}

	return "unknown failure occurred"
}

// Cause returns the underlying cause.
func (f Failure) Cause() error {
	return f.cause
}

func (f Failure) Unwrap() error {
	if f.IsEmpty() {
		return nil
	}
	return f.cause
}

// UnwrapFailure unwraps nested Failure types recursively.
func (f Failure) UnwrapFailure() Failure {
	if f.cause != nil {
		if inner, ok := errors.AsType[Failure](f.cause); ok {
			return inner.UnwrapFailure()
		}
	}

	return f
}

func (f Failure) IsPresent() bool {
	return f != emptyFailure
}

func (f Failure) IsEmpty() bool {
	return f == emptyFailure
}

// Is implements support for errors.Is matching.
func (f Failure) Is(target error) bool {
	if t, ok := errors.AsType[Failure](target); ok {
		return f.Equal(t)
	}
	return false
}

// OrThrow panics if the failure is present (Go's equivalent to throwing an exception).
func (f Failure) OrThrow() {
	if f.IsPresent() {
		panic(f)
	}
}

func (f Failure) IfPresent(action func(Failure)) {
	if action == nil {
		panic("action is nil")
	}

	if f.IsPresent() {
		action(f)
	}
}

func (f Failure) IfEmpty(runnable func()) {
	if runnable == nil {
		panic("runnable is nil")
	}

	if f.IsEmpty() {
		runnable()
	}
}

func (f Failure) InspectPresent(action func(Failure)) Failure {
	if action == nil {
		panic("action is nil")
	}

	if f.IsPresent() {
		action(f)
	}

	return f
}

func (f Failure) InspectEmpty(runnable func()) Failure {
	if runnable == nil {
		panic("runnable is nil")
	}

	if f.IsEmpty() {
		runnable()
	}

	return f
}

// Of executes a function/runnable and catches any panics or errors as a Failure.
func Of(runnable func()) (resFailure Failure) {
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
			msg := err.Error()
			if strings.TrimSpace(msg) == "" {
				msg = "An unexpected error occurred during runnable execution."
			}
			resFailure = Wrap(msg, err)
		}
	}()
	runnable()
	return emptyFailure
}

func (f Failure) Equal(other Failure) bool {
	// If both are empty failures
	if f.IsEmpty() && other.IsEmpty() {
		return true
	}

	// If only one is empty
	if f.IsEmpty() || other.IsEmpty() {
		return false
	}

	// Compare messages and check if the underlying causes match via errors.Is
	return f.message == other.message && errors.Is(f.cause, other.cause)
}

func (f Failure) String() string {
	if errors.Is(f, emptyFailure) {
		return "Failure.EMPTY"
	}

	result := fmt.Sprintf("Failure{message='%s'", f.Message())

	if f.cause != nil && !errors.Is(f, f.cause) {
		result += fmt.Sprintf(", cause=%T", f.cause)
		causeMessage := strings.TrimSpace(f.cause.Error())
		if causeMessage != "" {
			result += fmt.Sprintf("('%s')", causeMessage)
		}
	}

	result += "}"

	return result
}
