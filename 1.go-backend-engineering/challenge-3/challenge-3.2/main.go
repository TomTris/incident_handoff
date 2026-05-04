package main

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"
)

const numWorkers = 6
const timeOutInSecond = 4

var urls = []string{
	"https://google.com",
	"https://github.com",
	"https://go.dev",
	"https://pkg.go.dev",
	"https://cloudflare.com",
	"https://fastly.com",
	"https://stackoverflow.com",
	"https://reddit.com",
	"https://news.ycombinator.com",
	"https://gitlab.com",
	"https://bitbucket.org",
	"https://hub.docker.com",
	"https://kubernetes.io",
	"https://prometheus.io",
	// 6 URLs that will timeout
	"https://10.255.255.1",
	"https://10.255.255.2",
	"https://10.255.255.3",
	"https://10.255.255.1", // non-routable IP, hangs
	"https://192.0.2.1",    // TEST-NET, hangs
	"https://198.51.100.1", // TEST-NET-2, hangs
}

type Job struct {
	ID  int
	URL string
}

type Result struct {
	Job        Job
	WorkerID   int
	StatusCode int
	Duration   time.Duration
	Err        error
}

func isSuccessHttpCode(code int) bool {
	return code >= 200 && code < 300
}

func sendGetRequest(url string) (resp *http.Response, duration time.Duration, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeOutInSecond*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	timeStart := time.Now()
	resp, err = http.DefaultClient.Do(req)
	if err == nil {
		resp.Body.Close()
	}
	return resp, time.Since(timeStart), err
}

func worker(id int, resultsChan chan<- Result, jobsChan <-chan Job) {
	for job := range jobsChan {
		var result Result
		result.WorkerID = id
		result.Job = job
		resp, duration, err := sendGetRequest(job.URL)
		result.Err = err
		result.Duration = duration
		if err == nil {
			result.StatusCode = resp.StatusCode
		}
		resultsChan <- result
	}
}

func findMaxURLLength(urls []string) int {
	// Do not use sort to keep the order
	maxLength := 0
	for _, url := range urls {
		currentLen := len(url)
		if currentLen > maxLength {
			maxLength = currentLen
		}
	}
	return maxLength
}

func printInstantResult(maxURLLength int, result Result) {
	if result.Err == nil && isSuccessHttpCode(result.StatusCode) {
		fmt.Printf("[worker %d] %-2d ✅ %-*s → %d (%dms)\n",
			result.WorkerID, result.Job.ID, maxURLLength, result.Job.URL, result.StatusCode, result.Duration.Milliseconds())
	} else if result.Err == nil {
		fmt.Printf("[worker %d] %d ❌ %-*s → %d (%dms)\n",
			result.WorkerID, result.Job.ID, maxURLLength, result.Job.URL, result.StatusCode, result.Duration.Milliseconds())
	} else {
		fmt.Printf("[worker %d] %d ❌ %-*s → %s (%dms)\n",
			result.WorkerID, result.Job.ID, maxURLLength, result.Job.URL, result.Err, result.Duration.Milliseconds())
	}
}

func countHealthyResults(results []Result) int {
	cnt := 0
	for _, result := range results {
		if result.Err == nil && isSuccessHttpCode(result.StatusCode) {
			cnt++
		}
	}
	return cnt
}

func slowestHealthyResult(results []Result) Result {
	var slowestHealthyResult Result
	slowestDuration := time.Duration(0)
	for _, result := range results {
		if result.Err == nil && isSuccessHttpCode(result.StatusCode) && result.Duration > slowestDuration {
			slowestDuration = result.Duration
			slowestHealthyResult = result
		}
	}
	return slowestHealthyResult
}

func printSummary(results []Result, runtime int64) {
	sort.Slice(results, func(i, j int) bool {
		return (results[i].Duration < results[j].Duration)
	})
	numHealthyResults := countHealthyResults(results)
	slowest := slowestHealthyResult(results)
	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════")
	fmt.Println("                      SUMMARY")
	fmt.Println("══════════════════════════════════════════════════")
	fmt.Printf(" ✅  Healthy  (2xx)   :  %d\n", numHealthyResults)
	fmt.Printf(" ❌  Unreachable      :  %d\n", len(results)-numHealthyResults)
	fmt.Println("──────────────────────────────────────────────────")
	fmt.Printf(" Total                :  %d\n", len(results))
	fmt.Printf(" Fastest              :  %s (%dms)\n", results[0].Job.URL, results[0].Duration.Milliseconds())
	fmt.Printf(" Slowest (healthy)    :  %s (%dms)\n", slowest.Job.URL, slowest.Duration.Milliseconds())
	fmt.Printf(" Total runtime        :  %dms\n", runtime)
}

func main() {
	jobsChan := make(chan Job)
	resultsChan := make(chan Result)
	timeStart := time.Now()
	for i := 1; i <= numWorkers; i++ {
		go worker(i, resultsChan, jobsChan)
	}

	go func() {
		for idx, url := range urls {
			jobsChan <- Job{ID: idx, URL: url}
		}
		close(jobsChan)
	}()

	var results []Result
	maxURLLength := findMaxURLLength(urls)
	for len(results) != len(urls) {
		result := <-resultsChan
		results = append(results, result)
		printInstantResult(maxURLLength, result)
	}
	runtime := time.Since(timeStart).Milliseconds()
	printSummary(results, runtime)
}
