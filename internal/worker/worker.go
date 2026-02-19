package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"openradar/internal/config"
	"openradar/internal/db"
	"openradar/internal/domain"
	"openradar/internal/queue"
	"openradar/internal/scanner"
	"openradar/internal/server"

	"openradar/internal/scanner/detectors"

	"gorm.io/gorm"
)

var allowExt = map[string]struct{}{
	".env":  {},
	".md":   {},
	".txt":  {},
	".py":   {},
	".rs":   {},
	".yml":  {},
	".ts":   {},
	".js":   {},
	".yaml": {},
}

func hasTargetExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	_, ok := allowExt[ext]
	return ok
}

func runAllDetectors(src string, fileName string, scanJobID string, url string, DBtoSaveIn *gorm.DB) {
	for _, scanFunction := range detectors.AllDetectors {
		key, found, provider := scanFunction(src)
		if found && detectors.EnsureKeyIsntSpam(key) {
			log.Printf("Match found: %s\n", key)
			finding := domain.NewFinding(
				scanJobID,
				url,
				fileName,
				key,
				provider,
			)
			checkedFinding, err := db.GetFindingByKey(key, DBtoSaveIn)
			_ = checkedFinding

			if err != nil {
				if err := db.AddFinding(finding, DBtoSaveIn); err != nil {
					log.Printf("Failed to save finding for key %s: %v\n", key, err)
				}
			}
		}
	}
}

func cloneRepo(ctx context.Context, cloneURL string, dir string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--single-branch", cloneURL, dir)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func scanClonedFiles(dir string, scanJobID string, url string, DBtoSaveIn *gorm.DB, conf config.Config) error {
	var buf bytes.Buffer
	maxSize := int64(conf.Scanner.MaxFileSizeKB * 1024)

	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		if !hasTargetExt(d.Name()) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if info.Size() > maxSize {
			log.Printf("skipping file %s because it is too large (%d KB)", d.Name(), info.Size()/1024)
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}

		buf.Reset()
		_, err = io.Copy(&buf, f)
		f.Close()
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		runAllDetectors(buf.String(), relPath, scanJobID, url, DBtoSaveIn)
		return nil
	})
}

func Start(ctx context.Context, conf config.Config, DBtoSaveIn *gorm.DB, Hub *server.Hub) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-queue.JobQueue:
				log.Printf("Starting to process scan job %s for repository %s", job.ID, job.RepositoryURL)

				job.Status = domain.JobStatusInProgress
				repo, err := scanner.ScanRepo(context.Background(), job.RepositoryURL, conf.GitHub.Key)
				if err != nil {
					log.Printf("failed to scan repo %s: %v", job.RepositoryURL, err)
					continue
				}

				job.Status = domain.JobStatusCompleted
				job.UpdatedAt = time.Now()

				if repo.Size <= uint(conf.Scanner.MaxRepoSizeMB)*1000000 {
					dir, err := os.MkdirTemp("", "openradar-")
					if err != nil {
						log.Printf("failed to create temp dir: %v", err)
						continue
					}

					msg, err := json.Marshal(repo)
					if err != nil {
						log.Printf("Failed to send?")
					}
					Hub.Broadcast <- msg

					addedRepo := domain.NewRepository(
						job.ID,
						job.RepositoryURL,
					)

					cloneCtx, cloneCancel := context.WithTimeout(ctx, 60*time.Second)
					err = cloneRepo(cloneCtx, repo.Clone_Url, dir)
					cloneCancel()
					if err != nil {
						os.RemoveAll(dir)
						log.Printf("failed to clone repo %s: %v", job.RepositoryURL, err)
						continue
					}

					if err := scanClonedFiles(dir, job.ID, job.RepositoryURL, DBtoSaveIn, conf); err != nil {
						log.Printf("error while scanning files: %v", err)
					}

					os.RemoveAll(dir)

					existingRepo, err := db.GetRepositoryByName(job.RepositoryURL, DBtoSaveIn)
					_ = existingRepo

					if err != nil {
						if err := db.AddRepository(addedRepo, DBtoSaveIn); err != nil {
							log.Printf("Failed to save repository: %v", err)
						}
					} else {
						if err := db.UpdateRepository(addedRepo, DBtoSaveIn); err != nil {
							log.Printf("Failed to save repository: %v", err)
						}
					}
				}

				debug.FreeOSMemory()

				log.Printf("Finished processing scan job %s", repo.Url)
			}
		}
	}()
}
