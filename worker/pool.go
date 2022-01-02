package worker

import "sync"

func DoWithPool(poolSize int, works ...Work) []WorkResult {
	if len(works) == 0 || poolSize <= 0 {
		return []WorkResult{}
	}

	if len(works) == 1 {
		return []WorkResult{works[0]()}
	}

	return doConcurrentlyWithPool(poolSize, works)
}

func doConcurrentlyWithPool(poolSize int, works []Work) []WorkResult {
	worksChannel := generateWorksChannel(works)
	workResultsChannel := distributeWorks(poolSize, worksChannel)

	workResults := make([]WorkResult, 0, len(works))
	for workResult := range mergeWorkResults(workResultsChannel) {
		workResults = append(workResults, workResult)
	}

	return workResults
}

func generateWorksChannel(works []Work) <-chan Work {
	worksChannel := make(chan Work)

	go func() {
		for _, work := range works {
			worksChannel <- work
		}

		close(worksChannel)
	}()

	return worksChannel
}

func distributeWorks(poolSize int, works <-chan Work) []<-chan WorkResult {
	workResultsChannels := make([]<-chan WorkResult, 0, poolSize)

	for i := 0; i < poolSize; i++ {
		workResultsChannels = append(workResultsChannels, doWorksInPipeline(works))
	}

	return workResultsChannels
}

func doWorksInPipeline(works <-chan Work) <-chan WorkResult {
	workResultsChannel := make(chan WorkResult)

	go func() {
		for work := range works {
			workResultsChannel <- work()
		}

		close(workResultsChannel)
	}()

	return workResultsChannel
}

func mergeWorkResults(workResultsChannels []<-chan WorkResult) <-chan WorkResult {
	mergedWorkResultsChannel := make(chan WorkResult)
	var wg sync.WaitGroup
	wg.Add(len(workResultsChannels))

	drainWorkResultsChannel := func(workResultsChannel <-chan WorkResult) {
		for workResult := range workResultsChannel {
			mergedWorkResultsChannel <- workResult
		}

		wg.Done()
	}

	for _, workResultsChannel := range workResultsChannels {
		go drainWorkResultsChannel(workResultsChannel)
	}

	go func() {
		wg.Wait()
		close(mergedWorkResultsChannel)
	}()

	return mergedWorkResultsChannel
}
