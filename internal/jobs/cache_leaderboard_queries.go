package jobs

import (
	"openradar/internal/db"
	"openradar/internal/db/cache"
	"openradar/internal/domain"
	"sort"
	"strings"
	"time"
)

func extractUsername(repoName string) string {
	repoName = strings.TrimPrefix(repoName, "https://api.github.com/repos/")
	repoName = strings.TrimPrefix(repoName, "https://github.com/")
	parts := strings.Split(repoName, "/")
	if len(parts) >= 1 {
		return parts[0]
	}
	return repoName
}

func cache_leaderboard_query(jobContext JobContext) {
	database := jobContext.DB
	if database == nil {
		println("no db argument supplied")
		return
	}

	allFindings, err := db.GetAllFindings(database)
	if err != nil {
		println("error grabbing findings")
		return
	}

	userCounts := make(map[string]int)
	userRepos := make(map[string]string)
	for _, finding := range allFindings {
		username := extractUsername(finding.RepoName)
		userCounts[username]++
		userRepos[username] = finding.RepoName
	}

	var entries []domain.LeaderboardEntry
	for username, count := range userCounts {
		entries = append(entries, domain.LeaderboardEntry{
			Username: username,
			RepoName: userRepos[username],
			Leaks:    count,
			Avatar:   "https://github.com/" + username + ".png",
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Leaks > entries[j].Leaks
	})

	if len(entries) > 3 {
		entries = entries[:3]
	}

	cache.LeaderboardMu.Lock()
	cache.CachedLeaderboard = entries
	cache.LeaderboardMu.Unlock()
}

func init() {
	RegisterJob(
		Job{
			Name:     "Cache Database Queries for leaderboard",
			Func:     cache_leaderboard_query,
			Schedule: 5 * time.Minute,
		})
}
