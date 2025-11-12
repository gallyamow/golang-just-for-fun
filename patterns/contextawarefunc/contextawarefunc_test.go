package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

// @idiomatic: function is first-class-object
func TestContextAwareFunc1(t *testing.T) {
	t.Run("calls_fn", func(t *testing.T) {
		callsFn(t, ContextAwareRun1)
	})

	t.Run("cancelable", func(t *testing.T) {
		cancelable(t, ContextAwareRun1)
	})
}

func TestContextAwareFunc2(t *testing.T) {
	t.Run("calls_fn", func(t *testing.T) {
		callsFn(t, ContextAwareRun2)
	})

	t.Run("cancelable", func(t *testing.T) {
		cancelable(t, ContextAwareRun2)
	})
}

func TestContextAwareFunc3(t *testing.T) {
	t.Run("calls_fn", func(t *testing.T) {
		callsFn(t, ContextAwareRun3)
	})

	t.Run("cancelable", func(t *testing.T) {
		cancelable(t, ContextAwareRun3)
	})
}

func callsFn(t *testing.T, runner ContextAwareRunFunc[int]) {
	res, err := runner(t.Context(), testWorkBuilder(500*time.Millisecond))
	if err != nil {
		t.Fatalf("got error %v", err)
	}
	if res != 10 {
		t.Errorf("got %d, want 10", res)
	}
}

func cancelable(t *testing.T, runner ContextAwareRunFunc[int]) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	cancel()

	res, err := runner(ctx, testWorkBuilder(5*time.Second))
	if err == nil {
		t.Fatalf("got no error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got unexpected error %v", err)
	}

	if res != 0 {
		t.Errorf("got non zero value %d", res)
	}
}

func TestContextAwareChan(t *testing.T) {
	t.Run("calls_fn", func(t *testing.T) {
		ch1 := ContextAwareChan[int](t.Context(), testWorkBuilder(400*time.Millisecond))
		ch2 := ContextAwareChan[int](t.Context(), testWorkBuilder(200*time.Millisecond)) // first

		var resVal1, resVal2 = -1, -1

		for {
			select {
			case result, ok := <-ch1:
				if !ok {
					t.Errorf("ch1 closed")
				}

				if result.Err != nil {
					t.Fatalf("got error %v", result.Err)
				}

				resVal1 = result.Val
				if resVal1 != 10 {
					t.Errorf("got %d, want 10", resVal1)
				}

				if resVal2 == -1 {
					t.Errorf("ch1 earlier than ch2")
				}
			case result, ok := <-ch2:
				if !ok {
					t.Errorf("ch2 closed")
				}

				if result.Err != nil {
					t.Fatalf("got error %v", result.Err)
				}
				resVal2 = result.Val
				if resVal2 != 10 {
					t.Errorf("got %d, want 10", resVal2)
				}

				if resVal1 != -1 {
					t.Errorf("ch2 later than ch1")
				}
			case <-time.After(time.Second):
				t.Fatalf("timeout")
			}
		}
	})

	t.Run("cancelable", func(t *testing.T) {

	})
}

func testWorkBuilder(timeout time.Duration) workFunc[int] {
	return func() (int, error) {
		time.Sleep(timeout)
		return 10, nil
	}
}

// testWork долгая функция не обрабатывающая context
func testWork() (int, error) {
	time.Sleep(time.Second) // long-running operation
	return 10, nil
}
