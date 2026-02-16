package domain

import (
	"time"

	"github.com/google/uuid"
)

type ScanJobStatus string

const (
	JobStatusPending    ScanJobStatus = "pending"
	JobStatusInProgress ScanJobStatus = "in_progress"
	JobStatusCompleted  ScanJobStatus = "completed"
	JobStatusFailed     ScanJobStatus = "failed"
)

type ScanJob struct {
	ID            string        `json:"id"`
	RepositoryURL string        `json:"repository_url"`
	Status        ScanJobStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

func NewScanJob(repositoryURL string) *ScanJob {
	now := time.Now()
	return &ScanJob{
		ID:            uuid.New().String(),
		RepositoryURL: repositoryURL,
		Status:        JobStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
