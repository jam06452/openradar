package jobs

import (
	"openradar/internal/db"
	"time"
)

// Runs if repository has no findings & is old (> 24 hours)
func removeOldRepositoriesFunc(jobContext JobContext) {
	repos, err := db.GetAllRepositories(jobContext.DB)

	if err != nil {
		return
	}

	for _, repository := range repos {
		if time.Since(repository.LastUpdated) >= 24*time.Hour { // 24 hours

			findings, err := db.GetFindingsByRepo(jobContext.DB, repository.RepoName)
			if err != nil {
				return
			}

			if len(findings) == 0 { // No Findings!
				db.DeleteRepository(repository.ScanJobID, jobContext.DB) // Cleanup!
			}
		}
	}
}

func init() {
	RegisterJob(Job{
		Name:     "Remove old repositories",
		Func:     removeOldRepositoriesFunc,
		Schedule: 30 * time.Minute,
	})
}
