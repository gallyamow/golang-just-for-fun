package cache

import (
	"context"
	"golang.org/x/sync/errgroup"
)

// Для чего:
// Решает задачу как запустить работу в goroutine и корректно дождаться её результата, при этом не зависнуть,
// если контекст был отменён.
// Требования:
// 1) Принимает cancellable-контекст и функцию (функция именно не context-aware, иначе в данной обертке нет смысла).
// 2) Возвращает (Result, error)
//
// Нюанс любой реализации: так как fn не поддерживает контекст, то значит ее нельзя прервать на середине исполнения.
// Даже если contex cancelled, начавшая исполняться fn() и содержащая ее goroutine будет завершена.
//
// Основная идея:
// Мы запускаем fn() в отдельной goroutine, чтобы иметь возможность “соревноваться”
// между завершением fn() и сигналом ctx.Done().
//
// Это даёт нам контроль над временем ожидания — вызывающий код может вернуться
// раньше, если контекст отменён, не дожидаясь завершения fn().
//
// Важно понимать: поскольку Go не предоставляет механизм "прерывания"
// чужой goroutine (и fn не принимает контекст), сама fn() всё равно доработает до конца в фоне.
// Мы просто "забываем" о ней — результат уже не используется, но fn() продолжит
// выполняться, пока не завершится сама.

// ContextAwareRun1 реализация на основе done-channel и global переменных res и err.
func ContextAwareRun1[R any](ctx context.Context, fn WorkFunc[R]) (R, error) {
	var zero R

	// Этот способ работает и без этой проверки. Проверяем чтобы не запускать функцию даже один раз.
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}

	var res R
	var err error

	done := make(chan struct{})
	go func() {
		// @idiomatic: fire-and-forget pattern
		defer func() { _ = recover() }()
		defer close(done)
		res, err = fn()
	}()

	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-done:
		return res, err
	}
}

// ContextAwareRun2 реализация на основе Result-channel.
// @idiomatic: canonical way
func ContextAwareRun2[R any](ctx context.Context, fn WorkFunc[R]) (R, error) {
	var zero R

	// Этот способ работает и без этой проверки. Проверяем чтобы не запускать функцию даже один раз.
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}

	// @idiomatic: use a buffered channel to prevent goroutine leak
	resCh := make(chan Result[R], 1)
	go func() {
		// @idiomatic: fire-and-forget pattern
		defer func() { _ = recover() }()
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
func ContextAwareRun3[R any](ctx context.Context, fn WorkFunc[R]) (R, error) {
	var zero R
	var res R

	// Для этого способа важно:
	// Без этой проверки данный способ будет ждать таймаута если запущен с cancelled-контекстом
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		// @idiomatic: fire-and-forget pattern
		defer func() { _ = recover() }()
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

type ContextAwareRunFunc[R any] func(ctx context.Context, fn WorkFunc[R]) (R, error)
type WorkFunc[R any] func() (R, error)

type Result[R any] struct {
	Val R
	Err error
}

// ContextAwareChan реализация возвращает канал с результатом, что позволяет более гибко его использовать (например в
// select ожидающем готовности результата нескольких вызовов).
func ContextAwareChan[R any](ctx context.Context, fn WorkFunc[R]) <-chan Result[R] {
	var zero R

	out := make(chan Result[R], 1)

	// Обработка canceled контекста
	if ctx.Err() != nil {
		out <- Result[R]{Val: zero, Err: ctx.Err()}
		close(out)
		return out
	}

	go func() {
		// @idiomatic: fire-and-forget pattern
		defer func() { _ = recover() }()
		// @idiomatic: defer closing
		defer close(out)

		res, err := fn()

		// Вариант 1 — "Результат важнее контекста"
		// Хочу вернуть результат, если он успел рассчитаться, даже при отмене контекста
		/*
			select {
			case out <- Result[R]{Val: res, Err: err}:
			default:
				// ничего не делаем, контекст не проверяем
			}
		*/

		// Вариант 2 — "Контекст важнее результата":
		// Хочу, чтобы отмена контекста всегда "побеждала"
		select {
		case <-ctx.Done():
			out <- Result[R]{Val: zero, Err: ctx.Err()}
		default:
			// буфер точно свободен, можем сразу слать, тогда и результат будет без случайного выбора из 2 case
			out <- Result[R]{Val: res, Err: err}
		}
	}()

	return out
}

// ContextAwareWrapper реализация возвращающая context-aware версию функции (wraps).
// @idiomatic: modern way
func ContextAwareWrapper[R any](ctx context.Context, fn WorkFunc[R]) WorkFunc[R] {
	return func() (R, error) {
		return ContextAwareRun1(ctx, fn)
	}
}
