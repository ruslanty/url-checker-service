package checker

import (
	"context"
	"errors"
	"net/http"
	neturl "net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	Timeout     time.Duration
	Concurrency int
}

func ConfigFromEnv() Config {
	config := Config{
		Timeout:     5 * time.Second,
		Concurrency: 4,
	}

	if value, err := time.ParseDuration(os.Getenv("CHECK_TIMEOUT")); err == nil && value > 0 {
		config.Timeout = value
	}

	if value, err := strconv.Atoi(os.Getenv("CHECK_CONCURRENCY")); err == nil && value > 0 {
		config.Concurrency = value
	}

	return config
}

type Result struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	DurationMS int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

type Service struct {
	config Config
	client *http.Client
}

func New(config Config, client *http.Client) *Service {
	if client == nil {
		client = http.DefaultClient
	}

	return &Service{
		config: config,
		client: client,
	}
}

func (s *Service) Check(ctx context.Context, url string) Result {
	started := time.Now()

	// context and timeout limit one URL check
	checkCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	parsedURL, err := neturl.ParseRequestURI(url)

	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		return failure(url, started, "invalid URL")
	}

	request, err := http.NewRequestWithContext(checkCtx, http.MethodGet, url, nil)

	if err != nil {
		return failure(url, started, "invalid URL")
	}

	response, err := s.client.Do(request)

	if err != nil {
		return failure(url, started, requestError(checkCtx))
	}

	response.Body.Close()

	return Result{
		URL:        url,
		StatusCode: response.StatusCode,
		DurationMS: time.Since(started).Milliseconds(),
	}
}

func (s *Service) CheckMany(ctx context.Context, urls []string) []Result {
	results := make([]Result, len(urls))

	// channels distribute URL checks between the worker pool
	jobs := make(chan int)
	workers := s.config.Concurrency

	if workers > len(urls) {
		workers = len(urls)
	}

	if workers == 0 {
		return results
	}

	var wait sync.WaitGroup

	wait.Add(workers)

	// goroutines run worker pool with fixed concurrency limit
	for range workers {
		go func() {
			defer wait.Done()

			for index := range jobs {
				results[index] = s.Check(ctx, urls[index])
			}
		}()
	}

	for index := range urls {
		jobs <- index
	}

	close(jobs)

	wait.Wait()

	return results
}

func failure(url string, started time.Time, message string) Result {
	return Result{
		URL:        url,
		DurationMS: time.Since(started).Milliseconds(),
		Error:      message,
	}
}

func requestError(ctx context.Context) string {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "request timeout"
	}

	return "request failed"
}
