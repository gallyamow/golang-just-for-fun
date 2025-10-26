package throttler

import (
	"context"
	"time"
)

// TickerThrottler - работает на основе ticker.
// @idiomatic defer ordering
func TickerThrottler[T any](ctx context.Context, inputCh <-chan T, limit time.Duration) <-chan T {
	outputCh := make(chan T)

	// Выдает в канал ticker.C данные раз в limit времени.
	// При его остановке тикер не закрывает канал, чтобы не давать ожидающим частые тики (zero-values).
	// Таким образом если в одной goroutine чтение с канала остановленного ticker - то чтение блокируется.
	// (внутренняя реализация не на golang)
	//
	// Часто говорят что он может накапливать тики, но это не совсем так. Внутри у него буферизированный канал размером 1,
	// каждый новый тик отбрасывается если буфер уже заполнен.
	// Дело в том что если goroutine долго не читает из ticker.C, и потом начинает читать, она может получить тик сразу — потому что буфер уже содержит одно «накопленное» значение.
	ticker := time.NewTicker(limit)

	// Здесь используем именно указатель, так как нам отправленные значения "забывать", чтобы не отправить старое значение
	// при втором срабатывании ticker.
	var lastVal *T
	var first = true

	go func() {
		// Порядок важен: Будет выполняться с последнего до первого (так как стек LIFO).
		// Здесь:
		// - Сначала закрываем канал, чтобы сигнализировать, что больше данных не будет.
		// - Потом останавливаем тикер, чтобы освободить ресурсы после этого.
		// Неправильный порядок может приводить к send on closed channel: может успеть послать тик, даже после
		// ticker.Stop(), если событие уже было в очереди (Go не гарантирует немедленную остановку).
		defer ticker.Stop()
		defer close(outputCh)

		// Вычитываем все значения и пишем их output, но перед этим ждем ticker.
		// Вместо ticker можно и sleep(limit).
		// У sleep накапливается сдвиг, так как в него попадает время обработки других инструкций.
		for {
			// если первый вызов - то пишем результат сразу
			if first {
				select {
				case <-ctx.Done():
					return
				case val, ok := <-inputCh:
					if !ok {
						return
					}

					select {
					case <-ctx.Done():
						return
					case outputCh <- val:
						first = false
					}
				}
			} else {
				// Если вызов второй и далее - обновляем последнее lastVal значение.
				select {
				case <-ctx.Done():
					return
				case val, ok := <-inputCh:
					if !ok {
						return
					}

					if lastVal == nil {
						lastVal = &val
					} else {
						*lastVal = val
					}
				}

				// Если тик уже был при ожидании, то он сработает сразу
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					// Если значение есть - то отправляем
					// Запись осуществляем также с проверкой контекста.
					select {
					case <-ctx.Done():
						return
					case outputCh <- *lastVal:
						lastVal = nil
					}
				default:
					// Ждем в неблокирующем режиме, чтобы пропускать значения и дальше
				}
			}
		}
	}()

	return outputCh
}
