package main

import (
	"fmt"
)

func PipelineTest() {
	fmt.Println("PipelineTest")

	gen := func(count int) <-chan int {
		output := make(chan int)

		// запускам наполнение в goroutine, без нее мы не дойдем ни до close() ни до return, потому что нет читателей
		go func() {
			for i := range count {
				output <- i
			}

			// где создали - там и закрываем
			// закрытие именно тут после завершения всего цикла. Т.е. после закрытия входящего канала.
			close(output)
		}()

		return output
	}

	multiply := func(input <-chan int) <-chan int {
		output := make(chan int)

		go func() {
			for val := range input {
				output <- val * 2
			}

			close(output)
		}()

		return output
	}

	print := func(input <-chan int) {
		for val := range input {
			fmt.Println(val)
		}
	}

	// Создаем конвейер: gen -> multiply -> print
	print(multiply(gen(10)))

	fmt.Println("finished")
}

func main() {
	PipelineTest()
}
