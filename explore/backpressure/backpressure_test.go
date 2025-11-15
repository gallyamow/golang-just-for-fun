package zerovalues

import (
	"slices"
	"testing"
	"time"
)

const N = 5

// Backpressure (обратное давление) — это механизм управления потоком данных, когда производитель (producer)
// замедляется, если потребитель (consumer) не успевает обрабатывать данные.
//
// Для чего: Backpressure помогает «синхронизировать» скорость между производителем и потребителем.
//
// Способы реализации:
// 1. Ограничение через буферизованные каналы:
// Когда буфер заполнен, то producer будет ждать освобождения канала, а consumer сможет обрабатывать записи в своем темпе.
// 2. Явное подтверждение готовности:
// Consumer отправляет сигналы через отдельный канал, когда готов принять новые данные. Producer ждет сигнала.
// 3. Умный write c timeout на стороне producer:
// Producer пишет в канал через select { case out <- x; case time.After } и таки образом реагирует на переполнение канала.
// 4. Семафоры для ограничения параллелизма в producer для ограничения кол-ва отправок.
//
// Все способы будут сглаживать пиковую нагрузку, но не отменяет backpressure полностью.
func TestBackPressure(t *testing.T) {
	t.Run("non_buffered_finishes_despite_backpressure", func(t *testing.T) {
		maxDuration := 100 * time.Millisecond

		vals, elapsed := consumer(t,
			middleware(t,
				producer(t, -1, 5*time.Millisecond),
				-1,
				maxDuration,
			),
		)

		if len(vals) != N {
			t.Errorf("got %v, want %v", len(vals), N)
		}

		// we don't use
		want := []int{0, 1, 2, 3, 4}
		if !slices.Equal(vals, want) {
			t.Errorf("got %v, want %v", vals, want)
		}

		// processing time should be almost equal to the longest stage delay
		maxElapsed := maxDuration + maxDuration/10
		for _, el := range elapsed {
			if el > maxElapsed {
				t.Errorf("got %v, want %v", el, maxElapsed)
			}
		}
	})

	t.Run("buffered_finishes_despite_backpressure", func(t *testing.T) {
		maxDuration := 100 * time.Millisecond

		vals, elapsed := consumer(t,
			middleware(t,
				producer(t, 10, 5*time.Millisecond),
				10,
				maxDuration,
			),
		)

		if len(vals) != N {
			t.Errorf("got %v, want %v", len(vals), N)
		}

		// we don't use
		want := []int{0, 1, 2, 3, 4}
		if !slices.Equal(vals, want) {
			t.Errorf("got %v, want %v", vals, want)
		}

		// processing time should be almost equal to the longest stage delay
		maxElapsed := maxDuration + maxDuration/10
		for _, el := range elapsed {
			if el > maxElapsed {
				t.Errorf("got %v, want %v", el, maxElapsed)
			}
		}
	})

	t.Run("ack_implementation", func(t *testing.T) {
		maxDuration := 100 * time.Millisecond
		ackCh := make(chan struct{})

		vals, elapsed := ackConsumer(t,
			middleware(t,
				ackProducer(t, -1, 50*time.Millisecond, ackCh),
				-1,
				maxDuration,
			),
			ackCh,
		)

		close(ackCh)

		if len(vals) != N {
			t.Errorf("got %v, want %v", len(vals), N)
		}

		// we don't use
		want := []int{0, 1, 2, 3, 4}
		if !slices.Equal(vals, want) {
			t.Errorf("got %v, want %v", vals, want)
		}

		// processing time should be almost equal to the longest stage delay
		maxElapsed := maxDuration + maxDuration/10
		for _, el := range elapsed {
			if el > maxElapsed {
				t.Errorf("got %v, want %v", el, maxElapsed)
			}
		}
	})

	t.Run("smart_write_implementation", func(t *testing.T) {
		maxDuration := 100 * time.Millisecond

		vals, elapsed := consumer(t,
			middleware(t,
				smartProducer(t, -1, 50*time.Millisecond),
				-1,
				maxDuration,
			),
		)

		if len(vals) != N {
			t.Errorf("got %v, want %v", len(vals), N)
		}

		// we don't use
		want := []int{0, 1, 2, 3, 4}
		if !slices.Equal(vals, want) {
			t.Errorf("got %v, want %v", vals, want)
		}

		// processing time should be almost equal to the longest stage delay
		maxElapsed := maxDuration + maxDuration/10
		for _, el := range elapsed {
			if el > maxElapsed {
				t.Errorf("got %v, want %v", el, maxElapsed)
			}
		}
	})
}

// @idiomatic: returning a receive-only channel (the caller is not expected to write to it)
func producer(t *testing.T, bufferSize int, sleepDur time.Duration) <-chan int {
	outCh := buildChan(bufferSize)

	go func() {
		for v := range N {
			t.Logf("producer gonna send %d", v)
			outCh <- v
			t.Logf("producer sent %d", v)

			time.Sleep(sleepDur)
		}

		close(outCh)
	}()

	return outCh
}

// smartProducer пытается записать в канал только определенное время и таким образом реагирует на переполнение канала.
func smartProducer(t *testing.T, bufferSize int, sleepDur time.Duration) <-chan int {
	outCh := buildChan(bufferSize)

	go func() {
		for v := range N {
			t.Logf("smartProducer gonna send %d", v)

			sent := false

			// @idiomatic: retry-loop to repeat sending with condition (instead of break + label)
			for !sent {
				select {
				case outCh <- v:
					t.Logf("smartProducer sent %d", v)
					sent = true
				case <-time.After(sleepDur):
					// @idiomatic: classic way to drop some value
					t.Logf("consumer to slow, so we will repeat to send %d", v)
				}
			}
		}

		close(outCh)
	}()

	return outCh
}

// ackProducer генерирует новое значение только после получения ack-сигнала от ackConsumer.
func ackProducer(t *testing.T, bufferSize int, sleepDur time.Duration, ackCh <-chan struct{}) <-chan int {
	outCh := buildChan(bufferSize)

	go func() {
		for v := range N {
			t.Logf("smartProducer gonna send %d", v)
			outCh <- v
			t.Logf("smartProducer sent %d", v)

			// wait consumer to be ready
			// (разместил здесь, чтобы не слать стартовое значение)
			<-ackCh
		}

		close(outCh)
	}()

	return outCh
}

// ackConsumer сообщает ackConsumer что готов к следующему элементу после обработки текущего элемента.
func ackConsumer(t *testing.T, inCh <-chan int, ackCh chan<- struct{}) ([]int, []time.Duration) {
	var vals []int
	var elapsed []time.Duration

	last := time.Now()
	for v := range inCh {
		since := time.Since(last)
		vals = append(vals, v)
		elapsed = append(elapsed, since)

		t.Logf("consumer received %d after %v", v, elapsed)
		last = time.Now()

		// notifying consumer is ready
		ackCh <- struct{}{}
	}

	return vals, elapsed
}

// @idiomatic: using receive-only channel as input and receive-only channel as output
// @idiomatic: using goroutine to handle elements and return the channel instantly
func middleware(t *testing.T, inCh <-chan int, bufferSize int, sleepDur time.Duration) <-chan int {
	outCh := buildChan(bufferSize)

	go func() {
		for v := range inCh {
			t.Logf("middleware received %d, and gonna send", v)
			outCh <- v
			t.Logf("middleware sent %d", v)

			time.Sleep(sleepDur)
		}

		close(outCh)
	}()

	return outCh
}

func consumer(t *testing.T, inCh <-chan int) ([]int, []time.Duration) {
	var vals []int
	var elapsed []time.Duration

	last := time.Now()
	for v := range inCh {
		since := time.Since(last)
		vals = append(vals, v)
		elapsed = append(elapsed, since)

		t.Logf("consumer received %d after %v", v, elapsed)
		last = time.Now()
	}

	return vals, elapsed
}

func buildChan(bufferSize int) chan int {
	if bufferSize == -1 {
		return make(chan int)
	} else {
		return make(chan int, bufferSize)
	}
}
