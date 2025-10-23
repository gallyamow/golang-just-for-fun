package firstresult

import (
	"context"
	"testing"
)

func TestWay11(t *testing.T) {
	_, err := Way11(context.Background(), []string{"1", "2", "3"}, "key", resolvedGetter)

	if err != nil {
		t.Fatalf("error = %v", err)
	}
}

func TestWay12(t *testing.T) {
	_, err := Way12(context.Background(), []string{"1", "2", "3"}, "key", resolvedGetter)

	if err != nil {
		t.Fatalf("error = %v", err)
	}
}

func TestWay21(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		_, err := Way21(context.Background(), []string{"1", "2", "3"}, "key", RandGetter)

		if err != nil {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("failed", func(t *testing.T) {
		_, err := Way21(context.Background(), []string{"1", "2", "3"}, "key", FailedGetter)

		if err == nil {
			t.Fatalf("error = %v", err)
		}
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
