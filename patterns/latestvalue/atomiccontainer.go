package latestvalue

import (
	"sync/atomic"
)

// AtomicContainer - реализация контейнера на atomic-значении.
// Будет использоваться для benchmark c RWMutexContainer.
// @idiomatic atomic.Value instead of mutex
type AtomicContainer[T any] struct {
	value atomic.Value
}

func (c *AtomicContainer[T]) Set(val T) {
	c.value.Store(val)
}

func (c *AtomicContainer[T]) Get() T {
	// Вызов до store возвращает nil, наш интерфейс Container требует не nil.
	// Остальные реализации также возвращают zero value.
	// Альтернативная реализация могла бы быть реализована путем записи default значения в конструкторе.
	val := c.value.Load()

	if val == nil {
		var zero T
		return zero
	}

	// Типизация после, иначе нельзя сравнить с typed интерфейс будет не nil + компилятор не пропустит  (mismatched types T and untyped nil)
	return val.(T)
}
