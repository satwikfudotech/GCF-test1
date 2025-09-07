package loadgen

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	Success     bool          `json:"success"`
	Latency     time.Duration `json:"latency"`
	StatusCode  int           `json:"status_code"`
}

func Run(ctx context.Context, target string, totalReq int, concurrency int) []Result {
	results := make([]Result, 0, totalReq)
	client := &http.Client{Timeout: 10 * time.Second}

	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	for i := 0; i < totalReq; i++ {
		wg.Add(1)
		semaphore <- struct{}{} // limit concurrency

		go func() {
			defer wg.Done()
			start := time.Now()
			resp, err := client.Get(target)
			latency := time.Since(start)

			res := Result{}
			if err != nil {
				res = Result{Success: false, Latency: latency, StatusCode: 0}
			} else {
				res = Result{Success: resp.StatusCode == 200, Latency: latency, StatusCode: resp.StatusCode}
				resp.Body.Close()
			}

			mu.Lock()
			results = append(results, res)
			mu.Unlock()
			<-semaphore
		}()
	}
	wg.Wait()
	return results
}
