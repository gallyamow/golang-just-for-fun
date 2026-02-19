package goexit

import (
	"runtime"
	"testing"
)

func TestGoexit(t *testing.T) {
	t.Run("calls defers", func(t *testing.T) {
		done := make(chan struct{})
		var steps []string

		go func() {
			defer func() {
				steps = append(steps, "defer1")
			}()
			defer func() {
				steps = append(steps, "defer2")
			}()
			close(done)

			runtime.Goexit()
		}()

		<-done
		if len(steps) != 2 {
			t.Fatalf("expected 2 deferred calls, got %d (%v)", len(steps), steps)
		}
		if steps[0] != "defer2" || steps[1] != "defer1" {
			t.Fatalf("unexpected defer order: expected %v, got %v", []string{"defer2", "defer1"}, steps)
		}
	})

	t.Run("no panic", func(t *testing.T) {
		done := make(chan struct{})
		var p any

		go func() {
			defer func() {
				if r := recover(); r != nil {
					p = r
				}
			}()
			close(done)

			runtime.Goexit()
		}()

		<-done
		if p != nil {
			t.Fatalf("expected no panic thrown, got (%v)", p)
		}
	})

	t.Run("calls defers", func(t *testing.T) {
		done := make(chan struct{})
		var steps []string

		go func() {
			steps = append(steps, "level1")
			func() {
				steps = append(steps, "level2")
				func() {
					steps = append(steps, "level3")
					func() {
						steps = append(steps, "level4")
						func() {
							defer close(done)
							steps = append(steps, "level5")
							runtime.Goexit()
						}()
					}()
				}()
			}()
		}()

		<-done
		if len(steps) != 5 {
			t.Fatalf("expected 5 deferred calls, got %d (%v)", len(steps), steps)
		}
	})
}
