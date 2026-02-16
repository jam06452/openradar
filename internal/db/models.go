package db

import (
	"gorm.io/gorm"
)

// This represents the repository that has been scanned
type Repository struct {
	gorm.Model
	Url         string
	SizeInBytes uint

	ScrapedKeys []ScrapedKey
}

type ScrapedKey struct {
	gorm.Model

	RepositoryID uint
	KeyType      string
	KeyValue     string
}
