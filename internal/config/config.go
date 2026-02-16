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
		Addr         string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
	}

	Database struct {
		URL string
	}

	Redis struct {
		URL string
	}

	GitHub struct {
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
		log.Fatal(".env couldn't load!")
	}

	cfg.Env = getEnv("ENV", "development")

	cfg.HTTP.Addr = getEnv("HTTP_ADDR", ":8080")
	cfg.HTTP.ReadTimeout = mustDuration(getEnv("HTTP_READ_TIMEOUT", "10s"))
	cfg.HTTP.WriteTimeout = mustDuration(getEnv("HTTP_READ_TIMEOUT", "15s"))

	cfg.Database.URL = required("DATABASE_URL")
	cfg.Redis.URL = required("REDIS_URL")

	cfg.Scanner.MaxRepoSizeMB = mustInt(getEnv("SCAN_MAX_REPO_MB", "50"))
	cfg.Scanner.MaxFileSizeKB = mustInt(getEnv("SCAN_MAX_REPO_MB", "512"))
	cfg.Scanner.MaxConcurrentClones = mustInt(getEnv("SCAN_MAX_CONCURRENT", "5"))

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
