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

	t.Run("stops_on_error", func(t *testing.T) {
		stopsOnError(t, ContextAwareRun1)
	})
}

func TestContextAwareFunc2(t *testing.T) {
	t.Run("calls_fn", func(t *testing.T) {
		callsFn(t, ContextAwareRun2)
	})

	t.Run("cancelable", func(t *testing.T) {
		cancelable(t, ContextAwareRun2)
	})

	t.Run("stops_on_error", func(t *testing.T) {
		stopsOnError(t, ContextAwareRun2)
	})
}

func TestContextAwareFunc3(t *testing.T) {
	t.Run("calls_fn", func(t *testing.T) {
		callsFn(t, ContextAwareRun3)
	})

	t.Run("cancelable", func(t *testing.T) {
		cancelable(t, ContextAwareRun3)
	})

	t.Run("stops_on_error", func(t *testing.T) {
		stopsOnError(t, ContextAwareRun3)
	})
}

func callsFn(t *testing.T, runner ContextAwareRunFunc[int]) {
	res, err := runner(t.Context(), testWorkBuilder(10, 500*time.Millisecond))
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

	res, err := runner(ctx, testWorkBuilder(10, 5*time.Second))
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

func stopsOnError(t *testing.T, runner ContextAwareRunFunc[int]) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	res, err := runner(ctx, func() (int, error) {
		time.Sleep(time.Duration(100 * time.Millisecond))
		return 0, errors.New("some error")
	})

	if err == nil {
		t.Fatalf("got no error")
	}

	if res != 0 {
		t.Errorf("got non zero value %d", res)
	}
}

func TestContextAwareChan(t *testing.T) {
	t.Run("calls_fn", func(t *testing.T) {
		ch1 := ContextAwareChan[int](t.Context(), testWorkBuilder(20, 400*time.Millisecond))
		ch2 := ContextAwareChan[int](t.Context(), testWorkBuilder(10, 200*time.Millisecond)) // first

		var vals []int

		for {
			if len(vals) == 2 {
				break
			}
			select {
			case result, ok := <-ch1:
				if ok {
					vals = append(vals, result.Val)
				}
			case result, ok := <-ch2:
				if ok {
					vals = append(vals, result.Val)
				}
			case <-time.After(time.Second):
				t.Fatalf("timeout")
			}
		}

		var arr [2]int
		copy(arr[:], vals)

		if arr != [2]int{10, 20} {
			t.Errorf("got %v, want %v", arr, [2]int{10, 20})
		}
	})

	t.Run("cancelable", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		cancel()

		ch := ContextAwareChan[int](ctx, testWorkBuilder(10, 400*time.Millisecond))

		select {
		case <-ch:
			// Из-за того что select () resolved случайным образом, могут быть как успешный, так и canceled результаты
			// поэтому проверяем только то что он будет завершен раньше чем timeout.
			// (выбрал вариант 2, поэтому уже не совсем так)
		case <-time.After(time.Second):
			t.Fatalf("timeout")
		}

	})
}

// testWork долгая функция не обрабатывающая context
func testWorkBuilder(val int, timeout time.Duration) workFunc[int] {
	return func() (int, error) {
		time.Sleep(timeout)
		return val, nil
	}
}
