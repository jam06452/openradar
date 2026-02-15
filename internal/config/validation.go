package config

func validate(cfg Config) {
	if cfg.Scanner.MaxRepoSizeMB <= 0 {
		panic("invalid SCAN_MAX_REPO_MB")
	}
	if cfg.Scanner.MaxConcurrentClones <= 0 {
		panic("invalid SCAN_MAX_CONCURRENT")
	}
}
