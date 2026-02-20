package domain

type LeaderboardEntry struct {
	Username string `json:"username"`
	RepoName string `json:"repo_name"`
	Leaks    int    `json:"leaks"`
	Avatar   string `json:"avatar"`
}
