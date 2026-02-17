package worker

import (
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
		if found {
			log.Printf("Match found: %s\n", key)
			finding := domain.NewFinding(
				scanJobID,
				url,
				path.Name,
				key,
				provider,
			)
			checkedFinding, err := db.GetFindingByKey(key, DBtoSaveIn)
			_ = checkedFinding // golang xd

			if err != nil { // make sure key doesnt already have a entry in the db
				db.AddFinding(finding, DBtoSaveIn)
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
		defer r.Close()

		src, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		runAllDetectors(string(src), file, scanJobID, url, DBtoSaveIn)
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

				if repo.Size <= uint(conf.Scanner.MaxRepoSizeMB)*1000000 { // times 1000000x = mb
					dir, err := os.MkdirTemp("", "openradar-")
					if err != nil {
						log.Printf("failed to create temp dir: %v", err)
						continue
					}
					defer os.RemoveAll(dir)

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
						log.Printf("failed to clone repo %s: %v", job.RepositoryURL, err)
						continue
					}

					if err := loopThroughFiles(r, job.ID, job.RepositoryURL, DBtoSaveIn, conf); err != nil {
						log.Printf("error while looping through files: %v", err)
					}

					existingRepo, err := db.GetRepositoryByName(job.RepositoryURL, DBtoSaveIn)
					_ = existingRepo

					// Save/Update DB with repository
					if err != nil {
						// repo doesnt exist!
						if err := db.AddRepository(addedRepo, DBtoSaveIn); err != nil {
							log.Printf("Failed to save repository: %v", err)
						}
					} else {
						// repo already exists!
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
