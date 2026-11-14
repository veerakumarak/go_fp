package option

type Option[T any] struct {
	value   T
	present bool
}

func Some[T any](value T) Option[T] {
	return Option[T]{
		value:   value,
		present: true,
	}
}

func None[T any]() Option[T] {
	return Option[T]{
		present: false,
	}
}

func (o Option[T]) IsSome() bool {
	return o.present
}

func (o Option[T]) IsNone() bool {
	return !o.present
}

func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		panic("called Option.Unwrap() on a None value")
	}
	return o.value
}

func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.IsNone() {
		return defaultValue
	}
	return o.value
}

func (o Option[T]) UnwrapOrElse(f func() T) T {
	if o.IsNone() {
		return f()
	}
	return o.value
}

// Match provides a pattern-matching style handler for Some and None.
func (o Option[T]) Match(fSome func(T), fNone func()) {
	if o.present {
		fSome(o.value)
	} else {
		fNone()
	}
}

func (o Option[T]) IfSome(fSome func(T)) {
	if o.present {
		fSome(o.value)
	}
}
func (o Option[T]) IfNone(fNone func()) {
	if !o.present {
		fNone()
	}
}

// Standalone Functions - Go methods do not support adding new generic parameters to a method receiver. To transform an Option[T] to an Option[R], Map must be a package-level function
// Map applies a function to the contained value if Some, returning a new Option.
func Map[T any, R any](o Option[T], fn func(T) R) Option[R] {
	if !o.present {
		return None[R]()
	}
	return Some(fn(o.value))
}

// AndThen (FlatMap) chains computations that return an Option.
func AndThen[T any, R any](o Option[T], fn func(T) Option[R]) Option[R] {
	if !o.present {
		return None[R]()
	}
	return fn(o.value)
}
