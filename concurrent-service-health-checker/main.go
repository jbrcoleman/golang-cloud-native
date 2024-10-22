package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type ServiceHealth struct {
	Name      string
	URL       string
	Status    string
	Latency   time.Duration
	LastCheck time.Time
}

type HealthReport struct {
	Services  []ServiceHealth
	Timestamp time.Time
}

func checkService(service ServiceHealth, results chan<- ServiceHealth, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(service.URL)
	latency := time.Since(start)

	if err != nil {
		results <- ServiceHealth{
			Name:      service.Name,
			URL:       service.URL,
			Status:    "DOWN",
			Latency:   latency,
			LastCheck: time.Now(),
		}
		return
	}
	defer resp.Body.Close()

	status := "UP"
	if resp.StatusCode != http.StatusOK {
		status = fmt.Sprintf("DEGRADED (%d)", resp.StatusCode)
	}

	results <- ServiceHealth{
		Name:      service.Name,
		URL:       service.URL,
		Status:    status,
		Latency:   latency,
		LastCheck: time.Now(),
	}
}

func main() {
	services := []ServiceHealth{
		{Name: "GitHub API", URL: "https://api.github.com/status"},
		{Name: "Cloudflare DNS", URL: "https://1.1.1.1"},
		{Name: "Google DNS", URL: "https://8.8.8.8"},
		{Name: "HTTPBin", URL: "https://httpbin.org/status/200"},
		{Name: "JSONPlaceholder", URL: "https://jsonplaceholder.typicode.com/posts/1"},
	}

	results := make(chan ServiceHealth, len(services))
	var wg sync.WaitGroup

	// Start concurrent health checks
	for _, service := range services {
		wg.Add(1)
		go checkService(service, results, &wg)
	}

	// Wait in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var healthReport HealthReport
	healthReport.Timestamp = time.Now()

	for result := range results {
		healthReport.Services = append(healthReport.Services, result)
	}

	// Output results as JSON
	reportJSON, err := json.MarshalIndent(healthReport, "", "    ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(reportJSON))
}
