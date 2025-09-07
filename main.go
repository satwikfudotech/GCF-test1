package main

import (
	"encoding/json"
	"log"
	"net/http"

	"example.com/gcf-worker/loadgen"
)

type LoadTestRequest struct {
	TargetURL   string `json:"target_url"`
	Concurrency int    `json:"concurrency"`
	Requests    int    `json:"requests"`
}

type LoadTestResponse struct {
	TotalRequests int     `json:"total_requests"`
	SuccessCount  int     `json:"success_count"`
	FailCount     int     `json:"fail_count"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	ThroughputRps float64 `json:"throughput_rps"`
}

func LoadgenHTTP(w http.ResponseWriter, r *http.Request) {
	var req LoadTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Received load test request: %+v", req)

	res := loadgen.RunLoadTest(req.TargetURL, req.Concurrency, req.Requests)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoadTestResponse{
		TotalRequests: res.TotalRequests,
		SuccessCount:  res.SuccessCount,
		FailCount:     res.FailCount,
		AvgLatencyMs:  res.AvgLatencyMs,
		ThroughputRps: res.ThroughputRps,
	})
}

func main() {
	http.HandleFunc("/", LoadgenHTTP)
	port := "8080"
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
