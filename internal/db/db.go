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

	sqlDB.SetConnMaxIdleTime(10)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := db.AutoMigrate(Repository{}, ScrapedKey{}); err != nil {
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
		return fmt.Errorf("failed to create repository: %w", result.Error)
	}
	return nil
}

// Get All Repositories
func GetAllRepositories(db *gorm.DB) ([]domain.Repository, error) {
	var repos []domain.Repository
	result := db.Find(repos)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch repositories: %w", result.Error)
	}
	return repos, nil
}

// Get Findings By Repository
func GetFindingsByRepo(db *gorm.DB) error {
	var findings []domain.Finding
	result := db.Find(findings)
	if result.Error != nil {
		return fmt.Errorf("failed to fetch findings: %w", result.Error)
	}
	return nil
}

// Overwrite Repository
func UpdateRepository(repo *domain.Repository, db *gorm.DB) error {
	result := db.Save(repo)
	if result.Error != nil {
		return fmt.Errorf("failed to update repository: %w", result.Error)
	}
	return nil
}

// Overwrite Repository
func DeleteRepository(scanJobID string, db *gorm.DB) error {
	result := db.Where("scan_job_id = ?", scanJobID).Delete(&domain.Repository{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete repository: %w", result.Error)
	}
	return nil
}
