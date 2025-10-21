package firstresult

import (
	"context"
	"testing"
)

func TestWay11(t *testing.T) {
	val, err := Way11(context.Background(), []string{"1", "2", "3"}, "key", resolvedGetter)

	if err != nil {
		t.Fatalf("Way11() error = %v", err)
	}

	t.Logf("Way11() val = %v", val)
}

func TestWay12(t *testing.T) {
	val, err := Way12(context.Background(), []string{"1", "2", "3"}, "key", resolvedGetter)

	if err != nil {
		t.Fatalf("Way12() error = %v", err)
	}

	t.Logf("Way12() val = %v", val)
}

func TestWay21(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		val, err := Way21(context.Background(), []string{"1", "2", "3"}, "key", RandGetter)

		if err != nil {
			t.Fatalf("Way21() error = %v", err)
		}

		t.Logf("Way21() val = %v", val)
	})

	t.Run("failed", func(t *testing.T) {
		val, err := Way21(context.Background(), []string{"1", "2", "3"}, "key", FailedGetter)

		if err == nil {
			t.Fatalf("Way21() error = %v", err)
		}

		t.Logf("Way21() val = %v", val)
	})
}

func BenchmarkWay11(b *testing.B) {
	for b.Loop() {
		_, err := Way11(context.Background(), []string{"1", "2", "3"}, "key", resolvedGetter)
		if err != nil {
			b.Fatalf("Way11() error = %v", err)
		}
	}
}

func BenchmarkWay12(b *testing.B) {
	for b.Loop() {
		_, err := Way12(context.Background(), []string{"1", "2", "3"}, "key", resolvedGetter)

		if err != nil {
			b.Fatalf("Way12() error = %v", err)
		}
	}
}
