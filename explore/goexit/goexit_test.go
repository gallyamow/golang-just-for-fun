package goexit

import (
	"testing"
)

func TestGoexit(t *testing.T) {
	t.Run("calls defers", func(t *testing.T) {
		done := make(chan struct{})
		var steps []string

		func() {
			defer func() {
				steps = append(steps, "defer1")
			}()
			defer func() {
				steps = append(steps, "defer2")
			}()
			close(done)
		}()

		<-done
		if len(steps) != 2 {
			t.Fatalf("expected 2 deferred calls, got %d (%v)", len(steps), steps)
		}
		if steps[0] != "defer2" || steps[1] != "defer1" {
			t.Fatalf("unexpected defer order: expected %v, got %v", []string{"defer2", "defer1"}, steps)
		}

	})
}
