package firstresult

import (
	"context"
)

// Way12 solves problem 1
// - С buffered n-size каналом
func Way12(ctx context.Context, addresses []string, key string, getter ResolvedGetter) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	res := make(chan string, len(addresses))

	for _, addr := range addresses {
		go func(addr string) {
			// Такая проверка не совсем работает, но есть более правильный способ (см вариант 2)
			select {
			case <-ctx.Done():
			// Тут внутри case-handler есть блокирующаяся операция
			default:
				// В этом варианте почти всегда будет запрашиваться getResult, потому что это произойдет раньше чем, придет хотя бы один ответ.
				// Здесь оторванный от жизни кейс, где ошибка игнорируется.
				val := getter(ctx, addr, key)
				// Пишущие не будут блокироваться потому что канал достаточно большой.
				res <- val
			}
		}(addr)
	}

	// Благодаря defer cancel() все недошедшие до res <- val goroutines будут завершены (либо перейдут в запись,
	// так как switch case - при готовности нескольких выбирает случайный.)
	// А те goroutines те которые дошли публикации результата - смогут это сделать, так как канал с достаточным размером буфера.
	first := <-res

	// Закрывать канал не можем, потому что кто-то может писать.
	// Никогда не закрываем канал, в который пишут несколько goroutines, если есть риск, что кто-то ещё будет писать.

	// Можем сразу выходить.
	return first, nil
}
