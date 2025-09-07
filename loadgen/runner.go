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

func RunLoadTest(target string, concurrency int, requests int) Result {
    var mu sync.Mutex
    var wg sync.WaitGroup

    successes := 0
    failures := 0
    latencies := []float64{}

    startTime := time.Now()

    sem := make(chan struct{}, concurrency)

    for i := 0; i < requests; i++ {
        wg.Add(1)
        sem <- struct{}{}
        go func() {
            defer wg.Done()
            defer func() { <-sem }()

            start := time.Now()
            resp, err := http.Get(target)
            latency := time.Since(start).Milliseconds()

            mu.Lock()
            latencies = append(latencies, float64(latency))
            if err == nil && resp.StatusCode == http.StatusOK {
                successes++
            } else {
                failures++
            }
            mu.Unlock()
        }()
    }

    wg.Wait()
    totalDuration := time.Since(startTime).Seconds()

    avgLatency := 0.0
    for _, l := range latencies {
        avgLatency += l
    }
    if len(latencies) > 0 {
        avgLatency /= float64(len(latencies))
    }

    throughput := float64(requests) / totalDuration

    return Result{
        TotalRequests: requests,
        SuccessCount:  successes,
        FailCount:     failures,
        AvgLatencyMs:  avgLatency,
        ThroughputRps: throughput,
    }
}
