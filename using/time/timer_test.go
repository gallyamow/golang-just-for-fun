package time

import (
	"testing"
	"time"
)

func TestUsingTimer(t *testing.T) {
	t.Run("channel_is_not_closed_after_firing", func(t *testing.T) {
		tm := time.NewTimer(100 * time.Millisecond)

		select {
		case <-tm.C:
			t.Logf("timer.fired")
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("should have been fired already")
		}

		select {
		case <-tm.C:
			t.Fatalf("channel should not be closed")
		default:
			t.Logf("expected channel is not closed")
		}
	})

	t.Run("can_be_reused_by_reset", func(t *testing.T) {
		tm := time.NewTimer(50 * time.Millisecond)

		// остановка когда срок уже вышел, но данные из "С" не читали
		time.Sleep(150 * time.Millisecond)
		status := tm.Reset(100 * time.Millisecond)
		if !status {
			t.Fatal("got false, want true (as already stopped and not read)")
		}

		tm = time.NewTimer(50 * time.Millisecond)

		// остановка когда срок уже вышел, но данные из "С" читали
		time.Sleep(150 * time.Millisecond)
		<-tm.C
		status = tm.Reset(100 * time.Millisecond)
		if status {
			t.Fatal("got true, want false (as already stopped and read)")
		}

		// остановка пока срок не вышел
		time.Sleep(50 * time.Millisecond)
		status = tm.Reset(10 * time.Millisecond)
		if !status {
			t.Fatal("timer should not be expired")
		}
	})

	t.Run("afterfunc_convenient_usage", func(t *testing.T) {
		called := false
		tm := time.AfterFunc(100*time.Millisecond, func() {
			called = true
		})
		time.Sleep(50 * time.Millisecond)
		stopStatus := tm.Stop()
		if !stopStatus {
			t.Fatal("got false, want true (as stopped active timer)")
		}

		if called {
			t.Fatal("func was called, but should not have been because timer was stopped")
		}

		called = false
		tm = time.AfterFunc(50*time.Millisecond, func() {
			called = true
		})
		time.Sleep(100 * time.Millisecond)
		if !called {
			t.Fatal("func was not called, but should have been because timer was stopped")
		}
	})
}

func TestUsingTicker(t *testing.T) {
	// TODO
}
