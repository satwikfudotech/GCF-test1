package loadgen

import (
	"encoding/json"
	"log"
	"net/http"

	"example.com/gcf-worker/loadgen" // must match your go.mod
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
	if err := json.NewEncoder(w).Encode(LoadTestResponse{
		TotalRequests: res.TotalRequests,
		SuccessCount:  res.SuccessCount,
		FailCount:     res.FailCount,
		AvgLatencyMs:  res.AvgLatencyMs,
		ThroughputRps: res.ThroughputRps,
	}); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
