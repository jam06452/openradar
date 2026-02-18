package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"openradar/internal/domain"
	"openradar/internal/queue"
	"sync"
	"time"
)

var GITHUB_ENDPOINT = "https://api.github.com"

var sharedClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	},
}

type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Repo      Repo      `json:"repo"`
	CreatedAt time.Time `json:"created_at"`
	Payload   Payload   `json:"payload"`
}

type Repo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Payload struct {
	Ref  string `json:"ref"`
	Head string `json:"head"`
}

type ScannedRepository struct {
	Size      uint   `json:"size"`
	Url       string `json:"url"`
	Clone_Url string `json:"clone_url"`
}

var (
	recentlyScanned   = make(map[string]time.Time)
	recentlyScannedMu sync.Mutex
)

func cleanupRecentlyScanned() {
	recentlyScannedMu.Lock()
	defer recentlyScannedMu.Unlock()
	cutoff := time.Now().Add(-10 * time.Minute)
	for url, t := range recentlyScanned {
		if t.Before(cutoff) {
			delete(recentlyScanned, url)
		}
	}
}

func ScanJob(ctx context.Context, GITHUB_TOKEN string) ([]Event, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", GITHUB_ENDPOINT+"/events", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create req: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+GITHUB_TOKEN)

	res, err := sharedClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http call failed: %w", err)
	}

	defer res.Body.Close()

	var events []Event
	if err := json.NewDecoder(res.Body).Decode(&events); err != nil {
		log.Printf("json decode failed")
	}

	cleanupRecentlyScanned()

	recentlyScannedMu.Lock()
	for _, x := range events {
		if _, seen := recentlyScanned[x.Repo.URL]; seen {
			continue
		}
		recentlyScanned[x.Repo.URL] = time.Now()
		sampleJob := domain.NewScanJob(x.Repo.URL)
		select {
		case queue.JobQueue <- sampleJob:
		default:
			log.Printf("job queue full, dropping job for %s", x.Repo.URL)
		}
	}
	recentlyScannedMu.Unlock()

	return events, nil
}

func ScanRepo(ctx context.Context, URL string, GITHUB_TOKEN string) (ScannedRepository, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
	if err != nil {
		return ScannedRepository{}, fmt.Errorf("failed to create req: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+GITHUB_TOKEN)

	res, err := sharedClient.Do(req)
	if err != nil {
		return ScannedRepository{}, fmt.Errorf("http call failed: %w", err)
	}

	defer res.Body.Close()

	var repo ScannedRepository
	if err := json.NewDecoder(res.Body).Decode(&repo); err != nil {
		log.Printf("json decode failed")
	}

	return repo, nil
}
