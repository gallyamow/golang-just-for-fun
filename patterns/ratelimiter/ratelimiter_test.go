package ratelimiter

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	t.Run("allow", func(t *testing.T) {
		r := NewTokenBucket(10, 1)
		allow(t, r)
	})

	t.Run("cancel", func(t *testing.T) {
		r := NewTokenBucket(10, 1)
		cancelable(t, r)
	})

	t.Run("wait", func(t *testing.T) {
		r := NewTokenBucket(2, 1)
		wait(t, r)
	})
}

func TestLeakyBucket(t *testing.T) {
	t.Run("allow", func(t *testing.T) {
		r := NewLeakyBucket(10, time.Second)
		allow(t, r)
	})

	t.Run("cancel", func(t *testing.T) {
		r := NewLeakyBucket(10, time.Second)
		cancelable(t, r)
	})

	t.Run("wait", func(t *testing.T) {
		r := NewLeakyBucket(2, time.Second)
		wait(t, r)
	})
}

func allow(t *testing.T, r RateLimiter) {
	for i := 0; i < 10; i++ {
		if !r.Allow() {
			t.Errorf("must be allowed")
		}
	}

	if r.Allow() {
		t.Errorf("must not be allowed")
	}

	time.Sleep(1 * time.Second)
	if !r.Allow() {
		t.Errorf("must be allowed")
	}
}

func wait(t *testing.T, r RateLimiter) {
	// from capacity
	r.Wait(t.Context())
	r.Wait(t.Context())

	ts := time.Now()

	r.Wait(t.Context())
	elapsed := time.Since(ts).Milliseconds()

	if !(elapsed > 900 && elapsed < 1100) {
		t.Errorf("invalid period")
	}

	t.Log("expected")
}

func cancelable(t *testing.T, r RateLimiter) {
	ctx, cancel := context.WithCancel(t.Context())

	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	r.Wait(ctx)

	select {
	case <-time.After(10 * time.Millisecond):
		t.Log("expected")
	}
}
