package stackgrowth

import (
	"testing"
)

func TestConfirmForArray(t *testing.T) {
	p1, p2 := ConfirmForArray()

	if p1 == p2 {
		t.Fatalf("both pointers to one address")
	}

	t.Logf("expected: first %v, second %v", p1, p2)
}

func TestConfirmRandSizedSlice(t *testing.T) {
	p1, p2 := ConfirmRandSizedSlice()

	if p1 == p2 {
		t.Fatalf("both pointers to one address")
	}

	t.Logf("expected: first %v, second %v", p1, p2)
}

func TestConfirmNoChangesForEscapedValue(t *testing.T) {
	p1, p2 := ConfirmNoChangesForEscapedValue()

	if p1 != p2 {
		t.Fatalf("both pointers to one address")
	}

	t.Logf("expected: first %v, second %v", p1, p2)
}
