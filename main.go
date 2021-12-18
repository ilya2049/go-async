package main

import (
	"fmt"
	"net/http"
	"time"

	"go-async/web"
)

func main() {
	webWordCounter := web.NewWordCounter(&http.Client{
		Timeout: 3 * time.Second,
	})

	for _, result := range webWordCounter.CountWordInPages("div",
		"https://google.com/",
		"https://yandex.ru/",
		"https://unknown-page.x/",
		"https://www.kinopoisk.ru/",
		"https://www.wikipedia.org/",
		"https://go.dev/blog/",
	) {
		fmt.Println(result)
	}
}
