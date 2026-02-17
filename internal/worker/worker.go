package worker

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"openradar/internal/config"
	"openradar/internal/domain"
	"openradar/internal/queue"
	"openradar/internal/scanner"

	"openradar/internal/scanner/detectors"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
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

func runAllDetectors(src string, path *object.File) {
	for _, scanFunction := range detectors.AllDetectors {
		key, found := scanFunction(src)
		if found {
			log.Printf("Match found: %s\n", key)
		}
	}
}

func loopThroughFiles(repo *git.Repository) error {
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

		r, err := file.Reader()
		if err != nil {
			return err
		}
		defer r.Close()

		src, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		runAllDetectors(string(src), file)
		return nil
	})
	return err
}

func Start(conf config.Config) {
	go func() {
		for {
			job := queue.Dequeue()

			log.Printf("Starting to process scan job %s for repository %s", job.ID, job.RepositoryURL)

			job.Status = domain.JobStatusInProgress
			var repo = scanner.ScanRepo(job.RepositoryURL)

			job.Status = domain.JobStatusCompleted
			job.UpdatedAt = time.Now()

			log.Printf("Finished processing scan job %s", repo.Url)

			// 25 Mb
			if repo.Size <= uint(conf.Scanner.MaxRepoSizeMB)*1000000 { // times 1000000x = mb
				r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
					URL:      repo.Clone_Url,
					Progress: os.Stdout,
				})

				loopThroughFiles(r)

				print(r.Head())
				print(err)
			}
		}
	}()
	log.Println("Worker started and waiting for jobs :3")
}
