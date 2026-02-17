package scanner

import (
	"encoding/json"
	"log"
	"net/http"
	"openradar/internal/domain"
	"openradar/internal/queue"
	"time"
)

var GITHUB_ENDPOINT = "https://api.github.com"

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
	Size      uint   `json:"size"` // byte
	Url       string `json:"url"`
	Clone_Url string `json:"clone_url"`
}

// This scans the live push endpoint.
func ScanJob(GITHUB_TOKEN string) []Event {
	req, err := http.NewRequest("GET", GITHUB_ENDPOINT+"/events", nil) // live events api
	if err != nil {
		log.Fatalf("failed to create req: %s\n", err)
	}

	req.Header.Set("Authorization", "Bearer "+GITHUB_TOKEN)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("http call failed: %s\n", err) // replace with non fatal TODO
	}

	defer res.Body.Close()

	var events []Event
	if err := json.NewDecoder(res.Body).Decode(&events); err != nil {
		log.Printf("json decode failed")
	}

	for _, x := range events {
		sampleJob := domain.NewScanJob(x.Repo.URL)
		queue.Enqueue(sampleJob)
	}

	return events
}

// Scans repository
func ScanRepo(URL string, GITHUB_TOKEN string) ScannedRepository {
	req, err := http.NewRequest("GET", URL, nil) // live events api
	if err != nil {
		log.Fatalf("failed to create req: %s\n", err)
	}

	req.Header.Set("Authorization", "Bearer "+GITHUB_TOKEN)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("http call failed: %s\n", err) // replace with non fatal TODO
	}

	defer res.Body.Close()

	defer res.Body.Close()

	var repo ScannedRepository
	if err := json.NewDecoder(res.Body).Decode(&repo); err != nil {
		log.Printf("json decode failed")
	}

	return repo
}
