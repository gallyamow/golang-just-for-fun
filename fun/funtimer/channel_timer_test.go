package funtimer

import (
	"testing"
	"time"
)

func TestChannelTimer(t *testing.T) {
	t.Run("fires_after_duration", func(t *testing.T) {
		tm := NewTimer(50 * time.Millisecond)

		select {
		case <-tm.C:
			t.Log("expected")
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timer did not fire in time")
		}
	})

	t.Run("does_not_fire_after_stop", func(t *testing.T) {
		tm := NewTimer(50 * time.Millisecond)
		tm.Stop()

		select {
		case <-tm.C:
			t.Fatalf("timer fired despite being stopped")
		case <-time.After(100 * time.Millisecond):
			t.Log("expected")
		}
	})

	t.Run("reset_changes_timer_duration", func(t *testing.T) {
		tm := NewTimer(50 * time.Millisecond)

		reset := tm.Reset(100)
		if !reset {
			t.Fatal("Reset returned false before tick")
		}

		start := time.Now()
		select {
		case <-tm.C:
			if time.Since(start) < 90 {
				t.Fatal("Timer fired too early after reset")
			}
		case <-time.After(150 * time.Millisecond):
			t.Log("expected")
		}
	})

	t.Run("after_func_executes_func", func(t *testing.T) {
		done := make(chan struct{})
		tm := AfterFunc(50*time.Millisecond, func() {
			close(done)
		})

		select {
		case <-done:
			t.Log("expected")
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timer fired despite being stopped")
		}

		reset := tm.Reset(50)
		if reset {
			t.Fatal("Reset returned true after f executed")
		}
	})
}
