package domain

import (
	"time"

	"github.com/google/uuid"
)

type Finding struct {
	ID          string    `json:"id"`
	ScanJobID   string    `json:"scan_job_id"`
	RepoName    string    `json:"repo_name"`
	FilePath    string    `json:"file_path"`
	Line        int       `json:"line"`
	Severity    string    `json:"severity"` // e.g., "HIGH", "MEDIUM", "LOW", "INFO"
	Description string    `json:"description"`
	DetectedAt  time.Time `json:"detected_at"`
}

func NewFinding(scanJobID, repoName, filePath string, line int, severity, description string) *Finding {
	return &Finding{
		ID:          uuid.New().String(),
		ScanJobID:   scanJobID,
		RepoName:    repoName,
		FilePath:    filePath,
		Line:        line,
		Severity:    severity,
		Description: description,
		DetectedAt:  time.Now(),
	}
}
