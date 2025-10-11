package patterns

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// SelectWithTimeout: Использование select для ожидания на нескольких каналах с возможностью установки таймаутов или дедлайнов
func SelectWithTimeout() {
	input1 := make(chan int)
	input2 := make(chan string)
	input3 := make(chan float64)

	// Тут утечка горутин, потому что мы считаем 10, и цикл может незавершится в случае таймаутов.
	// 1) способ: Если сделать каналы буфферизированными то проблемы не будет, так как горутины завершаться, а оставшиеся в буфере значения будут удалены.
	// 2) способ: завершать через контекст
	// 3) способ - вычитывать до конца
	//
	// (если нет рекурсии то можно без var)
	// (можно на основе T)
	// (избавимся от утечки через контекст)
	anyWriter := func(ctx context.Context, out any) {
		defer func() {
			switch c := out.(type) {
			case chan int:
				close(c)
			case chan float64:
				close(c)
			case chan string:
				close(c)
			}
		}()

		for range 10 {
			// проверяем Done() в критических секциях, далее defer закроет канал
			select {
			case <-ctx.Done(): // Проверяем отмену
				return
			default: //default тут обязателен, иначе нижнее unreachable
			}

			val := rand.Intn(1000) + 1

			switch c := out.(type) {
			case chan int:
				select {
				case <-ctx.Done(): // Проверяем отмену и перед записью, потому что читателей уже нет
					return
				default:
					c <- val
				}
			case chan float64:
				select {
				case <-ctx.Done(): // Проверяем отмену и перед записью, потому что читателей уже нет
					return
				default:
					c <- float64(val)
				}
			case chan string:
				select {
				case <-ctx.Done(): // Проверяем отмену и перед записью, потому что читателей уже нет
					return
				default:
					c <- strconv.Itoa(val)
				}
			}
			time.Sleep(time.Duration(val) * time.Millisecond)
		}

	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go anyWriter(ctx, input1)
	go anyWriter(ctx, input2)
	go anyWriter(ctx, input3)

	timeout := 200 * time.Millisecond

	for range 10 {
		select {
		case <-time.After(timeout):
			fmt.Printf("timeout %v reached\n", timeout)
		case val := <-input1:
			fmt.Printf("value %v from %s\n", val, "input1")
		case val := <-input2:
			fmt.Printf("value %v from %s\n", val, "input2")
		case val := <-input3:
			fmt.Printf("value %v from %s\n", val, "input3")
		}
	}

	// просто так проверим нет ли непрочитанного значения
	select {
	case <-time.After(time.Duration(5000) * time.Microsecond):
		fmt.Printf("no values\n")
	case val := <-input1:
		fmt.Printf("value %v from %s\n", val, "input1")
	case val := <-input2:
		fmt.Printf("value %v from %s\n", val, "input2")
	case val := <-input3:
		fmt.Printf("value %v from %s\n", val, "input3")
	}

	fmt.Println("finished")
}
