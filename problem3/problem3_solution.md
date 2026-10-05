# Problem 3

The line `cnp := make(chan func(), 10)` creates a channel that holds functions, like a queue of jobs. The buffer size of 10 means up to ten jobs can wait in the channel; it does not mean ten workers are created. The loop `for i := 0; i < 4; i++` starts four goroutines using `go func()`. Each worker receives a function through `for f := range cnp` and runs it with `f()`. A job is received by only one worker, so although there are four workers, only one is needed for the single job in this example. This pattern can be used to send queued emails or process uploaded files while limiting how many jobs are worked on at once.

The line `cnp <- func() { fmt.Println("HERE1") }` submits a job but does not wait for it to finish. Since the channel has space, `main` continues to `fmt.Println("Hello")` and then returns. When `main` returns, the program exits without waiting for the other goroutines, so a worker may not get time to print `HERE1`. However, `HERE1` can still print if a worker runs before the program exits; it is not guaranteed to be missing every time. To guarantee it prints, `main` must wait for the job to finish, for example using a completion channel or a `sync.WaitGroup`.

## Using a WaitGroup

```go
func main() {
	cnp := make(chan func(), 10)
	var wg sync.WaitGroup

	wg.Add(4) // Count the workers before starting them.
	for i := 0; i < 4; i++ {
		go func() {
			defer wg.Done()
			for f := range cnp {
				f()
			}
		}()
	}

	cnp <- func() {
		fmt.Println("HERE1")
	}
	close(cnp) // No more jobs; let the workers finish their loops.
	wg.Wait()  // Keep main alive until all four workers finish.
	fmt.Println("Hello")
}
```

`wg.Add(4)` counts four workers. Each worker calls `wg.Done()` when it exits,
reducing the count by one. Closing `cnp` lets the workers finish after receiving
the queued job; without closing it, they would keep waiting for more jobs and
`wg.Wait()` would never finish. With this change, the output will always print `HERE1` followed by `Hello`