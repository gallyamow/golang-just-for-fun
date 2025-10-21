package firstresult

import (
	"context"
)

// Way21 solves problem 2
// - С buffered 1-size каналом
func Way21(ctx context.Context, addresses []string, key string, getter Getter) (string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Типизированный канал лучше чем any + switch case
	type result struct {
		val string
		err error
	}

	resCh := make(chan result, 1)

	for _, addr := range addresses {
		go func(addr string) {
			if ctx.Err() != nil {
				return
			}

			val, err := getter(ctx, addr, key)

			// Важно, что нет default, поэтому если канал заполнен, goroutine ждёт, пока освободится место.
			// Это гарантирует, что все результаты, включая успешные, будут переданы через канал, даже если канал временно заполнен.
			// Т.е. default тут нельзя, потому что можно пропустить какой-то successful результат, если канал в этот момент полон.
			// В way11 он нужен, чтобы не блокировать лишний раз + там нет ошибок и любой ответ ожидаемый.
			select {
			case <-ctx.Done():
			case resCh <- result{val, err}:
			}
		}(addr)
	}

	var lastErr error

	// Читаем именно столько раз сколько хостов. Просто range по resCh нельзя, так как канал не будет закрыт.
	// Поэтому читаем ровно столько сколько пишут.
	for range addresses {
		res := <-resCh
		if res.err == nil {
			// первый успешный результат
			// defer отменит остальные которые еще не начали писать, а начавшие писать не будут блокироваться, потому что
			// они в case, а не в case hanlder.
			return res.val, nil
		}
		lastErr = res.err
	}
	return "", lastErr
}
