package cache

import (
	"context"
	"golang.org/x/sync/errgroup"
)

// TODO: ContextAwareWrapper - еще лучше наверно

// Для чего:
// Решает задачу как запустить работу в goroutine и корректно дождаться её результата, при этом не зависнуть,
// если контекст был отменён.
// Требования:
// 1) Принимает cancellable-контекст и функцию (функция именно не context-aware, иначе в данной обертке нет смысла).
// 2) Возвращает (Result, error)

// ContextAwareRun1 реализация на основе done-channel.
func ContextAwareRun1[R any](ctx context.Context, fn workFunc[R]) (R, error) {
	var zero R

	// Этот способ работает и без этой проверки. Проверяем чтобы не запускать функцию даже один раз.
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}

	var res R
	var err error

	done := make(chan struct{})
	go func() {
		defer close(done)
		res, err = fn()
	}()

	select {
	case <-ctx.Done():
		var zero R
		return zero, ctx.Err()
	case <-done:
		return res, err
	}
}

// ContextAwareRun2 реализация на основе Result-channel.
// @idiomatic: canonical way
func ContextAwareRun2[R any](ctx context.Context, fn workFunc[R]) (R, error) {
	var zero R

	// Этот способ работает и без этой проверки. Проверяем чтобы не запускать функцию даже один раз.
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}

	resCh := make(chan Result[R], 1)
	go func() {
		defer close(resCh)
		res, err := fn()
		resCh <- Result[R]{res, err}
	}()

	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case result := <-resCh:
		return result.Val, result.Err
	}
}

// ContextAwareRun3 реализация на основе ErrGroup.
// @idiomatic: modern way
func ContextAwareRun3[R any](ctx context.Context, fn workFunc[R]) (R, error) {
	var zero R
	var res R

	// Для этого способа важно:
	// Без этой проверки данный способ будет ждать таймаута если запущен с cancelled-контекстом
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		r, err := fn()
		if err != nil {
			return err
		}

		res = r
		return nil
	})

	if err := g.Wait(); err != nil {
		return zero, err
	}

	return res, nil
}

type ContextAwareRunFunc[R any] func(ctx context.Context, fn workFunc[R]) (R, error)
type workFunc[R any] func() (R, error)

type Result[R any] struct {
	Val R
	Err error
}

// ContextAwareChan реализация возвращает канал с результатом, что позволяет более гибко его использовать (например в
// select ожидающем готовности результата нескольких вызовов).
func ContextAwareChan[R any](ctx context.Context, fn workFunc[R]) <-chan Result[R] {
	var zero R

	out := make(chan Result[R], 1)

	select {
	case <-ctx.Done():
		out <- Result[R]{Val: zero, Err: ctx.Err()}
	}

	return out
}
