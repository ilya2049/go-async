package worker

import "sync"

func Do(works ...Work) []WorkResult {
	if len(works) == 0 {
		return []WorkResult{}
	}

	if len(works) == 1 {
		return []WorkResult{works[0]()}
	}

	return doConcurrently(works)
}

func doConcurrently(works []Work) []WorkResult {
	var wg sync.WaitGroup
	resultsChannel := make(chan WorkResult, len(works))

	doWork := func(work func() WorkResult) {
		resultsChannel <- work()

		wg.Done()
	}

	wg.Add(len(works))

	for _, work := range works {
		go doWork(work)
	}

	wg.Wait()
	close(resultsChannel)

	results := make([]WorkResult, 0, len(works))
	for result := range resultsChannel {
		results = append(results, result)
	}

	return results
}
