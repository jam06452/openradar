package jobs

import (
	"openradar/internal/db"

	"time"

	"gorm.io/gorm"
)

// Runs if repository has no findings & is old (> 24 hours)

func RunJob(dbJob *gorm.DB) {
	repos, err := db.GetAllRepositories(dbJob)

	if err != nil {
		return
	}

	for _, repository := range repos {
		if time.Since(repository.LastUpdated) >= 24*time.Hour { // 24 hours

			findings, err := db.GetFindingsByRepo(dbJob, repository.RepoName)
			if err != nil {
				return
			}

			if len(findings) == 0 { // No Findings!
				db.DeleteRepository(repository.ScanJobID, dbJob) // Cleanup!
			}
		}
	}
}

func init() {
	AllJobs = append(AllJobs, RunJob)
}
