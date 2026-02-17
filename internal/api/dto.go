package api

import (
	"openradar/internal/domain"
)

type Finding struct {
	ScanJobId  string `json:"scan_job_id"`
	RepoUrl    string `json:"repo_name"`
	FilePath   string `json:"file_path"`
	DetectedAt string `json:"detected_at"`
	Key        string `json:"key"`
	Provider   string `json:"provider"`
}

type Repository struct {
	ScanJobId   string `json:"scan_job_id"`
	RepoName    string `json:"repo_name"`
	LastUpdated string `json:"last_updated"`
}

type PaginatedRepositories struct {
	Repositories []domain.Repository `json:"repositories"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	TotalCount   int64               `json:"total_count"`
	TotalPages   int                 `json:"total_pages"`
}

type PaginatedFindings struct {
	Findings   []domain.Finding `json:"findings"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalCount int64            `json:"total_count"`
	TotalPages int              `json:"total_pages"`
}
