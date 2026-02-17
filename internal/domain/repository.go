package domain

import (
	"time"
)

type Repository struct {
	ScanJobID  string    `json:"scan_job_id"`
	RepoName   string    `json:"repo_name"`
	DetectedAt time.Time `json:"detected_at"`
}

func NewRepository(scanJobID, repoName string) *Repository {
	return &Repository{
		ScanJobID:  scanJobID,
		RepoName:   repoName,
		DetectedAt: time.Now(),
	}
}
