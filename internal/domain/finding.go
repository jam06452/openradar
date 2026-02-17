package domain

import (
	"time"

	"github.com/google/uuid"
)

type Finding struct {
	ID         string    `json:"id"`
	ScanJobID  string    `json:"scan_job_id"`
	RepoName   string    `json:"repo_name"`
	FilePath   string    `json:"file_path"`
	DetectedAt time.Time `json:"detected_at"`
	Key        string    `json:"key"`
	Provider   string    `json:"provider"`
}

func NewFinding(scanJobID, repoName, filePath string, key string, provider string) *Finding {
	return &Finding{
		ID:         uuid.New().String(),
		ScanJobID:  scanJobID,
		RepoName:   repoName,
		FilePath:   filePath,
		Provider:   provider,
		DetectedAt: time.Now(),
		Key:        key,
	}
}
