package loadgen

import (
	"net/http"
	"sync"
	"time"
)

// Result stores the outcome of the test
type Result struct {
	TotalRequests int
	SuccessCount  int
	FailCount     int
	AvgLatencyMs  float64
	ThroughputRps float64
}

// RunLoadTest runs a simple load test
func RunLoadTest(target string, concurrency, requests int) Result {
	var wg sync.WaitGroup
	var mu sync.Mutex

	successCount := 0
	failCount := 0
	latencies := []float64{}

	start := time.Now()

	sem := make(chan struct{}, concurrency)

	for i := 0; i < requests; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			startTime := time.Now()
			resp, err := http.Get(target)
			latency := time.Since(startTime).Milliseconds()

			mu.Lock()
			defer mu.Unlock()
			latencies = append(latencies, float64(latency))
			if err == nil && resp.StatusCode == http.StatusOK {
				successCount++
			} else {
				failCount++
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start).Seconds()

	// calculate average latency
	var totalLatency float64
	for _, l := range latencies {
		totalLatency += l
	}
	avgLatency := totalLatency / float64(len(latencies))

	return Result{
		TotalRequests: requests,
		SuccessCount:  successCount,
		FailCount:     failCount,
		AvgLatencyMs:  avgLatency,
		ThroughputRps: float64(requests) / elapsed,
	}
}
