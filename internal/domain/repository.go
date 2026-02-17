package domain

import (
	"time"
)

type Repository struct {
	ScanJobID   string    `json:"scan_job_id"`
	RepoName    string    `json:"repo_name"`
	LastUpdated time.Time `json:"last_updated"`
}

func NewRepository(scanJobID, repoName string) *Repository {
	return &Repository{
		ScanJobID:   scanJobID,
		RepoName:    repoName,
		LastUpdated: time.Now(),
	}
}
