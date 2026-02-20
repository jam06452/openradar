package jobs

import (
	"openradar/internal/db"
	"sort"
	"strings"
	"sync"
	"time"
)

type LeaderboardEntry struct {
	Username string `json:"username"`
	RepoName string `json:"repo_name"`
	Leaks    int    `json:"leaks"`
	Avatar   string `json:"avatar"`
}

var (
	cachedLeaderboard []LeaderboardEntry
	leaderboardMu     sync.RWMutex
)

func GetCachedLeaderboard() []LeaderboardEntry {
	leaderboardMu.RLock()
	defer leaderboardMu.RUnlock()
	return cachedLeaderboard
}

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

	var entries []LeaderboardEntry
	for username, count := range userCounts {
		entries = append(entries, LeaderboardEntry{
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

	leaderboardMu.Lock()
	cachedLeaderboard = entries
	leaderboardMu.Unlock()
}

func init() {
	RegisterJob(
		Job{
			Name:     "Cache Database Queries for leaderboard",
			Func:     cache_leaderboard_query,
			Schedule: 5 * time.Minute,
		})
}
