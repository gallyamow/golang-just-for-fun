package fanout

import (
	"context"
	"testing"
)

func TestFanOut(t *testing.T) {
	t.Run("buffered_input", func(t *testing.T) {
		valsCnt := 10
		outputCnt := 5

		inputCh := make(chan int, 5)

		go func(valsCnt int) {
			for i := range valsCnt {
				inputCh <- i
			}
		}(valsCnt)

		var outputChs []chan<- int
		sum := 0

		for range outputCnt {
			ch := make(chan int)
			outputChs = append(outputChs, ch)

			go func(outputCh <-chan int) {
				for val := range outputCh {
					sum += val
				}
			}(ch)
		}

		for range outputCnt {
			ch := make(chan int, 4)
			outputChs = append(outputChs, ch)

			go func(outputCh chan int) {
				for val := range outputCh {
					sum += val
				}
			}(ch)
		}

		FanOut(context.Background(), inputCh, outputChs...)
	})
}
