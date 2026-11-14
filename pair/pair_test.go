package pair

import (
	"testing"
)

func TestNewAndGetters(t *testing.T) {
	p := New(42, "hello")

	if p.First() != 42 {
		t.Errorf("expected First() to be 42, got %v", p.First())
	}
	if p.Second() != "hello" {
		t.Errorf("expected Second() to be 'hello', got %v", p.Second())
	}
}

func TestString(t *testing.T) {
	p := New(10, true)
	expected := "Pair(10, true)"

	if p.String() != expected {
		t.Errorf("expected string representation '%s', got '%s'", expected, p.String())
	}
}

func TestEqual(t *testing.T) {
	p1 := New("key", 1)
	p2 := New("key", 1)
	p3 := New("key", 2)
	p4 := New("other", 1)

	// Test identical pairs
	if !Equal(p1, p2) {
		t.Errorf("expected p1 and p2 to be equal")
	}

	// Test different second element
	if Equal(p1, p3) {
		t.Errorf("expected p1 and p3 not to be equal")
	}

	// Test different first element
	if Equal(p1, p4) {
		t.Errorf("expected p1 and p4 not to be equal")
	}
}
