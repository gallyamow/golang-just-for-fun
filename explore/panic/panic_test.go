package panic

import (
	"strings"
	"sync"
	"testing"
)

// Panic — это сигнал об аварийной ситуации, когда программа не может продолжать обычное выполнение.
// Когда вызывается panic, Go делает следующее:
// 1) Останавливает текущую goroutine.
// 2) Выполняет все defer в обратном порядке их объявления (т.е. продолжаться ниже panic - не будет)
// 3) Если, паника не была поймана, то это goroutine завершается, если она является main-goroutine, то вся программа завершает выполнение
// с сообщением об ошибке и stack-trace.
//
// При использовании goroutines:
// Паника в одной goroutine не убивает другие goroutine. Если паника не поймана через recover, эта конкретная
// goroutine завершится с ошибкой, но программа продолжит работу, если это не main goroutine.
// При этом будет выведен stack-trace.
// Попытка поймать панику в recover main goroutine не сработает, ловить надо непосредственно в самой паникующей.
// Некоторые советуют ставить внутрь goroutines `defer func() { _ = recover() }()` в goroutines которых вы не ждете.
// (потому что в Go panic в goroutine не убивает родительскую goroutine, но может оставить незакрытые ресурсы.
// recover() предотвращает это) - но документации про это 1 страница на stackoverflow.
//
// Для чего:
// Паника — для критических, неожиданных ошибок, которые не имеют нормального пути обработки.
func TestPanic(t *testing.T) {
	t.Run("defer_calling", func(t *testing.T) {
		var res []string

		defer func() {
			res = append(res, "defer0")
		}()
		// final checker
		defer func() {
			// defer0 - не попадет, остальные defer попадут в обратном порядке.
			if strings.Join(res, ":") != "before:defer2:defer1:recover" {
				t.Errorf("got %v, want %v", strings.Join(res, ":"), "before:defer2:defer1:recover")
			}
		}()
		defer func() {
			if r := recover(); r != nil {
				res = append(res, "recover")
			}
		}()
		defer func() {
			res = append(res, "defer1")
		}()
		defer func() {
			res = append(res, "defer2")
		}()

		res = append(res, "before")

		panic("panic")

		// это не будет работать, даже после recover
		res = append(res, "after")
		if len(res) != 4 {
			t.Errorf("got %v, want %v", len(res), 4)
		}
	})

	t.Run("one_goroutine_fails", func(t *testing.T) {
		var res []string

		// recover из main не действует на панику в goroutine
		defer func() {
			if r := recover(); r != nil {
				res = append(res, "recover")
			}
		}()

		var wg sync.WaitGroup
		wg.Add(1)

		// @idiomatic: специально завернул в функцию чтобы
		go func() {
			defer wg.Done()
			res = append(res, "goroutine1")

			// чтобы тест прошел, завернут в анонимную функцию с recover. В самой goroutine специально не ловлю.
			func() {
				defer func() {
					if r := recover(); r != nil {
						res = append(res, "wrapped_recover")
					}
				}()
				panic("panic")
			}()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			res = append(res, "goroutine2")
		}()

		wg.Wait()

		// порядок может быть разным
		if strings.Join(res, ":") != "goroutine1:goroutine2:wrapped_recover" && strings.Join(res, ":") != "goroutine2:goroutine1:wrapped_recover" {
			t.Errorf("got %v, want %v", strings.Join(res, ":"), "?")
		}

		// доходит до конца, даже если не было recover
		t.Log("the main goroutine is complete despite panic in some of them")
	})

	t.Run("one_goroutine_fails_but_it_recovered", func(t *testing.T) {
		var res []string

		// recover из main не действует на панику в goroutine
		defer func() {
			if r := recover(); r != nil {
				res = append(res, "main_recover")
			}
		}()

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					res = append(res, "goroutine_recover")
				}
			}()
			defer wg.Done()
			res = append(res, "goroutine1")
			panic("panic")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			res = append(res, "goroutine2")
		}()

		wg.Wait()

		// порядок может быть разным
		if strings.Join(res, ":") != "goroutine1:goroutine2:goroutine_recover" && strings.Join(res, ":") != "goroutine2:goroutine1:goroutine_recover" {
			t.Errorf("got %v, want %v", strings.Join(res, ":"), "?")
		}

		// доходит до конца, даже если не было recover
		t.Log("the main goroutine is complete despite panic in some of them")
	})
}
