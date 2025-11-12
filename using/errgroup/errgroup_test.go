package errgroup

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"sync"
	"testing"
	"time"
)

// errgroup.Group — представляет собой обёртку над sync.WaitGroup, которая:
//
// Для чего он:
// 1) Автоматически собирает первую ошибку, если она произошла в какой-то goroutine.
// 2) Позволяет удобно отменять все goroutine через контекст (context.Context).
// 3) Позволяет безопасно дождаться завершения всех goroutines (g.Wait()).
//
// Как с ним работать:
// g, ctx := errgroup.WithContext(context.Background()) - создаем
// g.Go(func() error {}) - добавляем функцию в группу, функция будет запущена как goroutine
// g.SetLimit - устанавливаем лимит на количество goroutines
//
// Как реализован:
// Содержит:
// 1) WaitGroup - для дожидания завершения всех goroutines.
// 2) sem - чтобы регулировать параллельность запуска goroutines, по умолчанию этот канал не буферизирован и меняется на буферизированный в SetLimit.
// В случае если len(sem) != 0 паникует, значит действие происходит когда уже выполняется какая-то goroutine
// errOnce + err — безопасное хранение первой ошибки без гонки данных.
// cancel() используется для отмены связанного контекста, если какая-то goroutine возвращает ошибку. Вызывается также в Wait() после того как все отработали.
//
// Метод WithContext()
// Создает группу и контекст. Именно этот контекст надо обрабатывать внутри функции, и именно он будет отменен при первой ошибке.
// Метод Go():
// 1) Если есть лимит (g.sem), goroutine ждёт свободное место в семафоре.
// 2) Добавляем goroutine в WaitGroup.
// 3) Запускаем функцию f() в новой goroutine.
// 4) Если f() возвращает ошибку: то через errOnce.Do сохраняет эту ошибку (первую) и отменяет контекст.
// Метод done()
// Освобождает слот семафора и слот WaitGroup.
// Метод Wait()
// Ждёт все запущенные через Go goroutines (через wg.Wait()), Отменяет контекст после завершения, Возвращает первую ошибку, если она была.

func TestUsingErrorGroup(t *testing.T) {
	t.Run("sequential", func(t *testing.T) {
		var done []string
		var mu sync.Mutex

		task := func(ctx context.Context, name string, duration time.Duration) error {
			select {
			case <-time.After(duration):
				if name == "F" {
					t.Log("Failed:", name)
					return errors.New("failed")
				}

				mu.Lock()
				done = append(done, name)
				mu.Unlock()

				t.Log("Done:", name)
				return nil
			case <-ctx.Done():
				t.Log("Cancelled:", name)
				return ctx.Err()
			}
		}

		ctx := context.Background()
		g, ctx := errgroup.WithContext(ctx)

		tasks := []struct {
			name     string
			duration time.Duration
		}{
			{"A", 100 * time.Millisecond},
			{"B", 100 * time.Millisecond},
			{"C", 100 * time.Millisecond},
			{"D", 100 * time.Millisecond},
			{"F", 200 * time.Millisecond}, // failed
			{"G", 500 * time.Millisecond},
		}

		// Последовательное выполнение
		for _, ts := range tasks {
			ts := ts
			g.Go(func() error {
				return task(ctx, ts.name, ts.duration)
			})
		}

		// Ждем все + ошибку
		if err := g.Wait(); err == nil {
			t.Fatalf("want error")
		}

		// Завершены 4 первых несмотря на их длинный timeout
		if len(done) != 4 {
			t.Errorf("got %d, want %d", len(done), 4)
		}
	})

	t.Run("concurrently", func(t *testing.T) {
		var done []string
		var mu sync.Mutex

		task := func(ctx context.Context, name string, duration time.Duration) error {
			select {
			case <-time.After(duration):
				if name == "F" {
					t.Log("Failed:", name)
					return errors.New("failed")
				}

				mu.Lock()
				done = append(done, name)
				mu.Unlock()

				t.Log("Done:", name)
				return nil
			case <-ctx.Done():
				t.Log("Cancelled:", name)
				return ctx.Err()
			}
		}

		ctx := context.Background()
		g, ctx := errgroup.WithContext(ctx)
		g.SetLimit(6)

		tasks := []struct {
			name     string
			duration time.Duration
		}{
			{"A", 10 * time.Millisecond},
			{"B", 20 * time.Millisecond},
			{"C", 30 * time.Millisecond},
			{"D", 40 * time.Millisecond},
			{"F", 10 * time.Millisecond}, // failed
			{"G", 500 * time.Millisecond},
		}

		// Последовательное выполнение
		for _, ts := range tasks {
			ts := ts
			g.Go(func() error {
				return task(ctx, ts.name, ts.duration)
			})
		}

		// Ждем все + ошибку
		if err := g.Wait(); err == nil {
			t.Fatalf("want error")
		}

		// Завершен только A первый благодаря его коротком timeout
		if len(done) != 1 || done[0] != "A" {
			t.Errorf("got %d, want %d", len(done), 1)
		}
	})
}
