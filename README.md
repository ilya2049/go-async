## Worker pool (fan-in/fan-out)

``` go
func main() {
	pool := worker.NewPool(2)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	works := worker.GenerateSquares(ctx, 500*time.Millisecond)
	results := pool.Run(ctx, works)

	fmt.Println("1)", <-results)
	fmt.Println("2)", <-results)

	// Give the time to generate and do more works that won't be received.
	time.Sleep(2 * time.Second)

	fmt.Println("cancel context!")
	cancel()

	// Try to drain a result channel before the cancel broadcast will reach the runners.
	for result := range results {
		fmt.Println("*)", result)
	}

	// Give the time all runners to print about closing.
	time.Sleep(1 * time.Second)
}
```