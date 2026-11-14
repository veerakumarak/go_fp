package pair

import "fmt"

type Pair[A any, B any] struct {
	a A
	b B
}

func New[A any, B any](a A, b B) Pair[A, B] {
	return Pair[A, B]{
		a: a,
		b: b,
	}
}

func (p Pair[A, B]) First() A {
	return p.a
}

func (p Pair[A, B]) Second() B {
	return p.b
}

func (p Pair[A, B]) String() string {
	return fmt.Sprintf("Pair(%v, %v)", p.a, p.b)
}

// helper methods
func Equal[A comparable, B comparable](p1 Pair[A, B], p2 Pair[A, B]) bool {
	return p1.a == p2.a && p1.b == p2.b
}
