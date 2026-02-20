package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env string

	HTTP struct {
		Port string
	}

	Database struct {
		URL string
	}

	GitHub struct {
		Key string
	}

	Scanner struct {
		MaxRepoSizeMB       int
		MaxFileSizeKB       int
		MaxConcurrentClones int
	}
}

func Load() Config {
	var cfg Config

	err := godotenv.Load()
	if err != nil {
		log.Println(".env couldn't load! using env variables.")
	}

	cfg.Env = getEnv("ENV", "development")

	cfg.Database.URL = required("DATABASE_URL")

	cfg.Scanner.MaxRepoSizeMB = mustInt(getEnv("SCAN_MAX_REPO_MB", "50"))
	cfg.Scanner.MaxFileSizeKB = mustInt(getEnv("SCAN_MAX_FILE_KB", "2048"))
	cfg.Scanner.MaxConcurrentClones = mustInt(getEnv("SCAN_MAX_CONCURRENT", "5"))

	cfg.GitHub.Key = required("GITHUB_TOKEN")

	cfg.HTTP.Port = required("PORT")

	validate(cfg)
	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback

}

func required(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic("missing required env: " + key)
	}
	return val
}

func mustInt(val string) int {
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}
	return i
}

func mustDuration(val string) time.Duration {
	d, err := time.ParseDuration(val)
	if err != nil {
		panic(err)
	}
	return d
}
