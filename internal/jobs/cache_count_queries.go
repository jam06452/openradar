// yes this caches queries
// yes its very cool
// yes i love you :heart:S

package jobs

import (
	"fmt"
	"time"

	"openradar/internal/db/cache"
	"openradar/internal/domain"

	"gorm.io/gorm"
)

func get_finding(db *gorm.DB) (int64, error) {
	var totalCount int64
	if err := db.Model(&domain.Finding{}).Count(&totalCount).Error; err != nil {
		return 0, fmt.Errorf("error counting findings: %w", err)
	}
	return totalCount, nil
}

func get_repository(db *gorm.DB) (int64, error) {
	var totalCount int64
	if err := db.Model(&domain.Repository{}).Count(&totalCount).Error; err != nil {
		return 0, fmt.Errorf("error counting repositories: %w", err)
	}
	return totalCount, nil
}

func cache_count_queries(jobContext JobContext) {
	db := jobContext.DB

	count, err := get_finding(db)
	if err == nil {
		cache.FindingsCount = count
	}

	count2, err := get_repository(db)
	if err == nil {
		cache.RepositoriesCount = count2
	}
}

func init() {
	RegisterJob(
		Job{
			Name:     "Cache queries for count",
			Func:     cache_count_queries,
			Schedule: 30 * time.Second,
		})
}
