package db

import (
	"fmt"

	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"openradar/internal/domain"
)

func New(url string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgress: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(&domain.Repository{}, &domain.Finding{}); err != nil {
		return nil, fmt.Errorf("auto-migrate failed: %w", err)
	}

	return db, nil
}

// New repository entry
func AddRepository(repo *domain.Repository, db *gorm.DB) error {
	result := db.Create(repo)
	if result.Error != nil {
		return fmt.Errorf("failed to create repository: %w", result.Error)
	}
	return nil
}

// New finding entry
func AddFinding(finding *domain.Finding, db *gorm.DB) error {
	result := db.Create(finding)
	if result.Error != nil {
		return fmt.Errorf("failed to create finding: %w", result.Error)
	}
	return nil
}

// Get All Repositories
func GetAllRepositories(db *gorm.DB) ([]domain.Repository, error) {
	var repos []domain.Repository
	result := db.Find(&repos)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", result.Error)
	}
	return repos, nil
}

// Get Repository by Name
func GetRepositoryByName(id string, db *gorm.DB) (*domain.Repository, error) {
	var repo domain.Repository
	result := db.First(&repo, "repo_name = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository not found")
		}
		return nil, fmt.Errorf("failed to fetch repository: %w", result.Error)
	}
	return &repo, nil
}

// Get Finding by Key
func GetFindingByKey(key string, db *gorm.DB) (*domain.Finding, error) {
	var find domain.Finding
	result := db.First(&find, "key = ?", key)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("key not found")
		}
		return nil, fmt.Errorf("failed to fetch key: %w", result.Error)
	}
	return &find, nil
}

// Get Findings By Repository
func GetFindingsByRepo(db *gorm.DB, repoName string) ([]domain.Finding, error) {
	var findings []domain.Finding
	result := db.Where("repo_name = ?", repoName).Find(&findings)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch findings: %w", result.Error)
	}
	return findings, nil
}

// Overwrite Repository
func UpdateRepository(repo *domain.Repository, db *gorm.DB) error {
	result := db.Save(repo)
	if result.Error != nil {
		return fmt.Errorf("failed to update repository: %w", result.Error)
	}
	return nil
}

// Remove Repository
func DeleteRepository(scanJobID string, db *gorm.DB) error {
	result := db.Where("scan_job_id = ?", scanJobID).Delete(&domain.Repository{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete repository: %w", result.Error)
	}
	return nil
}
