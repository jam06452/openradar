package worker

import (
	"bytes"
	"context"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"openradar/internal/config"
	"openradar/internal/db"
	"openradar/internal/domain"
	"openradar/internal/queue"
	"openradar/internal/scanner"

	"openradar/internal/scanner/detectors"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"os"

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

func runAllDetectors(src string, path *object.File, scanJobID string, url string, DBtoSaveIn *gorm.DB) {
	for _, scanFunction := range detectors.AllDetectors {
		key, found, provider := scanFunction(src)
		if found && detectors.EnsureKeyIsntSpam(key) {
			log.Printf("Match found: %s\n", key)
			finding := domain.NewFinding(
				scanJobID,
				url,
				path.Name,
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

func loopThroughFiles(repo *git.Repository, scanJobID string, url string, DBtoSaveIn *gorm.DB, conf config.Config) error {
	ref, err := repo.Head()
	if err != nil {
		return err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	tree.Files().ForEach(func(file *object.File) error {
		if !hasTargetExt(file.Name) {
			return nil
		}

		if file.Size > int64(conf.Scanner.MaxFileSizeKB*1024) {
			log.Printf("skipping file %s because it is too large (%d KB)", file.Name, file.Size/1024)
			return nil
		}

		r, err := file.Reader()
		if err != nil {
			return err
		}

		buf.Reset()
		_, err = io.Copy(&buf, r)
		r.Close()
		if err != nil {
			return err
		}

		runAllDetectors(buf.String(), file, scanJobID, url, DBtoSaveIn)
		return nil
	})
	return err
}

func Start(ctx context.Context, conf config.Config, DBtoSaveIn *gorm.DB) {
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

					addedRepo := domain.NewRepository(
						job.ID,
						job.RepositoryURL,
					)

					r, err := git.PlainClone(dir, false, &git.CloneOptions{
						URL:      repo.Clone_Url,
						Progress: nil,
						Depth:    1,
					})
					if err != nil {
						os.RemoveAll(dir)
						log.Printf("failed to clone repo %s: %v", job.RepositoryURL, err)
						continue
					}

					if err := loopThroughFiles(r, job.ID, job.RepositoryURL, DBtoSaveIn, conf); err != nil {
						log.Printf("error while looping through files: %v", err)
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

				log.Printf("Finished processing scan job %s", repo.Url)
			}
		}
	}()
}
