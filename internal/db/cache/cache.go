package cache

import (
	"openradar/internal/domain"
	"sync"
)

var (
	FindingsCount     int64
	RepositoriesCount int64
)

var (
	CachedLeaderboard []domain.LeaderboardEntry
	LeaderboardMu     sync.RWMutex
)

func GetCachedLeaderboard() []domain.LeaderboardEntry {
	LeaderboardMu.RLock()
	defer LeaderboardMu.RUnlock()
	return CachedLeaderboard
}
