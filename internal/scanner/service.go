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
	Size uint   `json:"size"` // byte
	Url  string `json:"url"`
}

// This scans the live push endpoint.
func ScanJob() []Event {
	res, err := http.Get(GITHUB_ENDPOINT + "/events") // live events api
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
func ScanRepo(URL string) ScannedRepository {
	res, err := http.Get(URL)
	if err != nil {
		log.Fatalf("http call failed: %s\n", err)
	}

	defer res.Body.Close()

	var repo ScannedRepository
	if err := json.NewDecoder(res.Body).Decode(&repo); err != nil {
		log.Printf("json decode failed")
	}

	return repo
}
