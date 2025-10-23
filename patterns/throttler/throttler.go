package main

import (
	"context"
	"time"
)

// Throttler - она создаёт output-канал, который получает элементы из input-канал,
// но передаёт их не чаще, чем раз в limit времени, то есть регулирует пропускную способность (throughput).
//
// Требования:
//   - принимает input-канал для чтения и возвращать output-канал
//   - output должен выдавать значение не чаще чем в limit
//   - реагирует на отмену через контекст
//
// @idiomatic channel throttler
// @idiomatic defer ordering
func Throttler[T any](ctx context.Context, inputCh <-chan T, limit time.Duration) <-chan T {
	outputCh := make(chan T)

	// Выдает в канал ticker.C данные раз в limit времени.
	// При его остановке тикер не закрывает канал, чтобы не давать ожидающим частые тики (zero-values).
	// Таким образом если в одной goroutine чтение с канала остановленного ticker - то чтение блокируется.
	// (внутренняя реализация не на golang)
	ticker := time.NewTicker(limit)

	go func() {
		// Порядок важен: Будет выполняться с последнего до первого (так как стек LIFO).
		// Здесь:
		// - Сначала закрываем канал, чтобы сигнализировать, что "всё, больше данных не будет.
		// - Потом останавливаем тикер, чтобы освободить ресурсы после этого.
		// Неправильный порядок может приводить к send on closed channel.
		defer ticker.Stop()
		defer close(outputCh)

		// Вычитываем все значения и пишем их output, но перед этим ждем ticker.
		// Вместо ticker можно и sleep(limit).
		// У sleep накапливается сдвиг, так как в него попадает время обработки других инструкций.
		for i := range inputCh {
			<-ticker.C
			outputCh <- i
		}
	}()

	return outputCh
}
