package api

import (
	"fmt"
	"math"
	"openradar/internal/domain"
	"time"

	"gorm.io/gorm"
)

func GetLatestFindings(page int, pageSize int, provider string, minAge string, dbToGrabFrom *gorm.DB) (*PaginatedFindings, error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if pageSize < 1 || pageSize > 100 {
		return nil, fmt.Errorf("page_size must be between 1 and 100")
	}

	duration, err := time.ParseDuration(minAge)
	if err != nil {
		return nil, fmt.Errorf("invalid minAge format")
	}

	if duration < 0 || duration > 365*24*time.Hour {
		return nil, fmt.Errorf("minAge must be between 0 and 1 year")
	}

	cutOffTime := time.Now().Add(-duration)

	var findings []domain.Finding
	var totalCount int64

	query := dbToGrabFrom.Model(&domain.Finding{}).Where("detected_at >= ?", cutOffTime)

	if provider != "*" {
		validProviders := map[string]bool{
			"anthropic":  true,
			"cerebras":   true,
			"google":     true,
			"groq":       true,
			"mistral":    true,
			"openai":     true,
			"openrouter": true,
			"xai":        true,
		}
		if !validProviders[provider] {
			return nil, fmt.Errorf("invalid provider: %s", provider)
		}
		query = query.Where("provider = ?", provider)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("error counting findings: %w", err)
	}

	offset := (page - 1) * pageSize

	result := query.
		Order("detected_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&findings)

	if result.Error != nil {
		return nil, fmt.Errorf("error fetching findings: %w", result.Error)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &PaginatedFindings{
		Findings:   findings,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

func GetRepositoryInfo(repo_url string, dbToGrabFrom *gorm.DB) ([]domain.Repository, error) {
	var repository []domain.Repository

	if repo_url == "" {
		return nil, fmt.Errorf("repo_url cannot be empty")
	}

	query := dbToGrabFrom.Where("repo_name = ?", repo_url)

	result := query.Find(&repository)
	if result.Error != nil {
		return nil, fmt.Errorf("error fetching repository: %w", result.Error)
	}

	return repository, nil
}

func GetFindingsFromRepository(repo_url string, page int, pageSize int, dbToGrabFrom *gorm.DB) (*PaginatedFindings, error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if pageSize < 1 || pageSize > 100 {
		return nil, fmt.Errorf("page_size must be between 1 and 100")
	}
	if repo_url == "" {
		return nil, fmt.Errorf("repo_url cannot be empty")
	}

	var findings []domain.Finding
	var totalCount int64

	query := dbToGrabFrom.Model(&domain.Finding{}).Where("repo_name = ?", repo_url)

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("error counting findings: %w", err)
	}

	offset := (page - 1) * pageSize

	result := query.
		Order("detected_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&findings)

	if result.Error != nil {
		return nil, fmt.Errorf("error fetching findings: %w", result.Error)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &PaginatedFindings{
		Findings:   findings,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

func GetAllRepositories(page int, pageSize int, dbToGrabFrom *gorm.DB) (*PaginatedRepositories, error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if pageSize < 1 || pageSize > 100 {
		return nil, fmt.Errorf("page_size must be between 1 and 100")
	}

	var repositories []domain.Repository
	var totalCount int64

	if err := dbToGrabFrom.Model(&domain.Repository{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("error counting repositories: %w", err)
	}

	offset := (page - 1) * pageSize

	result := dbToGrabFrom.
		Order("last_updated DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&repositories)

	if result.Error != nil {
		return nil, fmt.Errorf("error fetching repositories: %w", result.Error)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &PaginatedRepositories{
		Repositories: repositories,
		Page:         page,
		PageSize:     pageSize,
		TotalCount:   totalCount,
		TotalPages:   totalPages,
	}, nil
}

func GetFindingsCount(db *gorm.DB) (int64, error) {
	var totalCount int64
	if err := db.Model(&domain.Finding{}).Count(&totalCount).Error; err != nil {
		return 0, fmt.Errorf("error counting findings: %w", err)
	}
	return totalCount, nil
}
