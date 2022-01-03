package worker

import (
	"context"
	"fmt"
	"time"
)

func GenerateSquares(ctx context.Context, interval time.Duration) chan Work {
	works := make(chan Work)

	go func() {
		i := 1

		for {
			select {
			case <-ctx.Done():
				close(works)
				fmt.Println("generator stopped")

				return

			case works <- newWork(i):
				fmt.Println("work added: ", i)
				i++
				time.Sleep(interval)
			}
		}
	}()

	return works
}

func newWork(i int) Work {
	return func() WorkResult { return i * i }
}
