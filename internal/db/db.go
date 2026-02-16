package db

import (
	"fmt"

	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
