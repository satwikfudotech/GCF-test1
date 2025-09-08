package loadgen

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LoadRequest struct {
	TargetURL   string `json:"target_url"`
	Requests    int    `json:"requests"`
	Concurrency int    `json:"concurrency"`
}

type Result struct {
	Success    bool          `json:"success"`
	Latency    time.Duration `json:"latency"`
	StatusCode int           `json:"status_code"`
}

// Cloud Function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	var req LoadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Limit max batch per worker
	const maxBatch = 1000
	const maxConcurrency = 100

	total := req.Requests
	results := []Result{}

	for total > 0 {
		batch := maxBatch
		if total < batch {
			batch = total
		}

		concurrency := req.Concurrency
		if concurrency > maxConcurrency {
			concurrency = maxConcurrency
		}

		fmt.Printf("Running batch of %d requests with concurrency %d\n", batch, concurrency)
		batchResults := runBatch(req.TargetURL, batch, concurrency)
		results = append(results, batchResults...)

		total -= batch
	}

	// Return all results to master
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func runBatch(target string, total, concurrency int) []Result {
	results := make([]Result, 0, total)
	sem := make(chan struct{}, concurrency)

	for i := 0; i < total; i++ {
		sem <- struct{}{}
		go func() {
			start := time.Now()
			resp, err := http.Get(target)
			latency := time.Since(start)

			res := Result{Latency: latency}
			if err == nil {
				res.StatusCode = resp.StatusCode
				res.Success = resp.StatusCode == 200
				resp.Body.Close()
			}
			results = append(results, res)
			<-sem
		}()
	}

	// wait
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
	return results
}
