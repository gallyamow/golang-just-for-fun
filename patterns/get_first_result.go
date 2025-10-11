package patterns

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Паттерн для получения первого готового результата.
func GetFirstResultTest() {
	get := func(ctx context.Context, address string, key string) (string, error) {
		fmt.Printf("going get %s from %s\n", key, address)

		time.Sleep(time.Duration(rand.Intn(20)+1) * time.Millisecond)
		// if rand.Float32() < 0.3 {
		// 	return "", errors.New("failed")
		// }

		return fmt.Sprintf("value of %s from address %s", key, address), nil
	}

	// упрощенная с буфферизированным каналом: тогда можно обойтись без done и waitGroup
	getFirstVal := func(ctx context.Context, addresses []string, key string) (string, error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		res := make(chan string, len(addresses))

		for _, addr := range addresses {
			go func(addr string) {
				select {
				case <-ctx.Done():
				default:
					// вариант через select case res <- не подходит, так как в таком случае все равно результат будет запрашиваться
					val, _ := get(ctx, addr, key)
					res <- val
				}
			}(addr)
		}

		// WaitGroup не нужен, можно не ждать, cancel() в конце отменит выполнение недошедших до res <- val горутин
		// А те которые дошли - смогут записать в буфферизированный канал.

		first := <-res
		fmt.Println("got firstly val", first)

		// Дочитывать остаток не надо, канал будет уничтожен при выходе из функции.

		return first, nil
	}

	for range 1 {
		val, err := getFirstVal(context.Background(), []string{"host1", "host2", "host3"}, "key1")
		fmt.Printf("val %v err %v\n", val, err)
	}

	fmt.Println("finished")
}
