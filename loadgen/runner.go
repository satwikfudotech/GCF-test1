package loadgen

import (
	"net/http"
	"sync"
	"time"
)

type Result struct {
	TotalRequests int
	SuccessCount  int
	FailCount     int
	AvgLatencyMs  float64
	ThroughputRps float64
}

// RunLoadTest executes requests in batches with concurrency limits
func RunLoadTest(target string, concurrency, totalRequests int) Result {
	const maxBatch = 1000
	const maxConcurrency = 100

	var allLatencies []time.Duration
	success := 0
	fail := 0

	remaining := totalRequests
	start := time.Now()

	for remaining > 0 {
		batch := maxBatch
		if remaining < batch {
			batch = remaining
		}

		c := concurrency
		if c > maxConcurrency {
			c = maxConcurrency
		}

		bSuccess, bFail, latencies := runBatch(target, batch, c)
		success += bSuccess
		fail += bFail
		allLatencies = append(allLatencies, latencies...)

		remaining -= batch
	}

	totalTime := time.Since(start).Seconds()

	// compute avg latency
	var sum time.Duration
	for _, l := range allLatencies {
		sum += l
	}
	avgLatency := float64(sum.Milliseconds()) / float64(len(allLatencies))

	return Result{
		TotalRequests: totalRequests,
		SuccessCount:  success,
		FailCount:     fail,
		AvgLatencyMs:  avgLatency,
		ThroughputRps: float64(totalRequests) / totalTime,
	}
}

func runBatch(target string, total, concurrency int) (int, int, []time.Duration) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)
	success := 0
	fail := 0
	var mu sync.Mutex
	latencies := make([]time.Duration, 0, total)

	for i := 0; i < total; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			start := time.Now()
			resp, err := http.Get(target)
			latency := time.Since(start)

			mu.Lock()
			latencies = append(latencies, latency)
			if err == nil && resp.StatusCode == 200 {
				success++
			} else {
				fail++
			}
			mu.Unlock()

			if resp != nil {
				resp.Body.Close()
			}
			<-sem
		}()
	}

	wg.Wait()
	return success, fail, latencies
}
