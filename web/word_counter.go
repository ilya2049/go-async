package web

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go-async/worker"
)

type WordCountingResult struct {
	PageURL   string
	WordCount int

	Err error
}

func (r WordCountingResult) String() string {
	if r.Err != nil {
		return fmt.Sprintf("%s (error): %s", r.PageURL, r.Err)
	}

	return fmt.Sprintf("%s: %d", r.PageURL, r.WordCount)
}

func NewWordCounter(anHTTPClient httpClient) *WordCounter {
	return &WordCounter{
		anHTTPClient: anHTTPClient,
	}
}

type WordCounter struct {
	anHTTPClient httpClient
}

func (s *WordCounter) CountWordInPages(word string, pageURLs ...string) []WordCountingResult {
	works := make([]worker.Work, 0, len(pageURLs))

	for _, pageURL := range pageURLs {
		works = append(works, s.newCountingWordsWork(pageURL, word))
	}

	return unpackWorkResults(worker.Do(works...))
}

func (s *WordCounter) newCountingWordsWork(pageURL, word string) worker.Work {
	return func() worker.WorkResult {
		wordsCount, err := s.countWordsInPageContents(pageURL, word)
		if err != nil {
			return WordCountingResult{
				PageURL: pageURL,
				Err:     err,
			}
		}

		return WordCountingResult{
			PageURL:   pageURL,
			WordCount: wordsCount,
		}
	}
}

var errResponseHasNotOKStatus = errors.New("response has not 200 status")

func (s *WordCounter) countWordsInPageContents(pageURL, word string) (int, error) {
	response, err := s.anHTTPClient.Get(pageURL)
	if err != nil {
		return 0, fmt.Errorf("failed to get page contents: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return 0, errResponseHasNotOKStatus
	}

	pageContents, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read page contents: %w", err)
	}

	defer response.Body.Close()

	return strings.Count(string(pageContents), word), nil
}

func unpackWorkResults(workResults []worker.WorkResult) []WordCountingResult {
	wordCountingResults := make([]WordCountingResult, 0, len(workResults))

	for _, workResult := range workResults {
		if wordCountingResult, ok := workResult.(WordCountingResult); ok {
			wordCountingResults = append(wordCountingResults, wordCountingResult)
		}
	}

	return wordCountingResults
}
