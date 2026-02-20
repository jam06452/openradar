package domain

import (
	"time"
)

type Repository struct {
	ScanJobID   string    `json:"scan_job_id"`
	RepoName    string    `json:"repo_name" gorm:"primaryKey"`
	LastUpdated time.Time `json:"last_updated"`
}

type PaginatedRepositories struct {
	Repositories []Repository `json:"repositories"`
	Page         int          `json:"page"`
	PageSize     int          `json:"page_size"`
	TotalCount   int64        `json:"total_count"`
	TotalPages   int          `json:"total_pages"`
}

func NewRepository(scanJobID, repoName string) *Repository {
	return &Repository{
		ScanJobID:   scanJobID,
		RepoName:    repoName,
		LastUpdated: time.Now(),
	}
}
