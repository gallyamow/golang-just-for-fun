package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	t.Run("allow", func(t *testing.T) {
		tb := NewTokenBucket(10, 1)
		for i := 0; i < 10; i++ {
			if !tb.Allow() {
				t.Errorf("must be allowed")
			}
		}

		if tb.Allow() {
			t.Errorf("must not be allowed")
		}

		time.Sleep(1 * time.Second)
		if !tb.Allow() {
			t.Errorf("must be allowed")
		}
	})

	t.Run("cancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())

		tb := NewTokenBucket(10, 1)

		go func() {
			time.Sleep(5 * time.Millisecond)
			cancel()
		}()

		tb.Wait(ctx)

		select {
		case <-time.After(10 * time.Millisecond):
			t.Log("expected")
		}
	})

	t.Run("wait", func(t *testing.T) {
		tb := NewTokenBucket(2, 1)

		// from capacity
		tb.Wait(t.Context())
		tb.Wait(t.Context())

		ts := time.Now()

		tb.Wait(t.Context())
		elapsed := time.Since(ts).Milliseconds()

		if !(elapsed > 900 && elapsed < 1100) {
			t.Errorf("invalid period")
		}

		t.Log("expected")
	})
}
