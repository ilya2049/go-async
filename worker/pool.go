package worker

import (
	"context"
	"fmt"
	"sync"
)

type Pool struct {
	size int
}

func NewPool(size int) *Pool {
	if size <= 0 {
		panic("worker: invalid pool size: must be greater than zero")
	}

	return &Pool{
		size: size,
	}
}

func (p *Pool) Run(ctx context.Context, worksChan <-chan Work) <-chan WorkResult {
	resultChannels := make([]<-chan WorkResult, 0, p.size)

	for i := 0; i < p.size; i++ {
		resultChannels = append(resultChannels, p.newRunner(ctx, worksChan))
	}

	return p.mergeRunners(ctx, resultChannels)
}

func (*Pool) newRunner(ctx context.Context, worksChan <-chan Work) <-chan WorkResult {
	resultChannel := make(chan WorkResult)

	go func() {
		defer func() {
			fmt.Println("runner stopped")
			close(resultChannel)
		}()

		var result WorkResult
		for work := range worksChan {
			result = work()

			select {
			case <-ctx.Done():
				return

			case resultChannel <- result:
				fmt.Println("work is done:", result)
			}
		}
	}()

	return resultChannel
}

func (*Pool) mergeRunners(ctx context.Context, resultsChannels []<-chan WorkResult) <-chan WorkResult {
	mergedResultChannel := make(chan WorkResult)

	var wg sync.WaitGroup
	wg.Add(len(resultsChannels))

	drain := func(resultChannel <-chan WorkResult) {
		defer func() {
			fmt.Println("drain runner stopped")
			wg.Done()
		}()

		for result := range resultChannel {
			select {
			case <-ctx.Done():
				return

			case mergedResultChannel <- result:
				fmt.Println("result merged:", result)
			}
		}
	}

	for _, resultChannel := range resultsChannels {
		go drain(resultChannel)
	}

	go func() {
		wg.Wait()
		close(mergedResultChannel)
	}()

	return mergedResultChannel
}
