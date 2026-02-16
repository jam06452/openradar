package worker

import (
	"log"
	"time"

	"openradar/internal/domain"
	"openradar/internal/queue"
	"openradar/internal/scanner"
)

func Start() {
	go func() {
		for {
			job := queue.Dequeue()

			log.Printf("Starting to process scan job %s for repository %s", job.ID, job.RepositoryURL)

			job.Status = domain.JobStatusInProgress
			var repo = scanner.ScanRepo(job.RepositoryURL)

			job.Status = domain.JobStatusCompleted
			job.UpdatedAt = time.Now()

			log.Printf("Finished processing scan job %s", repo.Url)
			log.Printf("Size: %d", repo.Size)
		}
	}()
	log.Println("Worker started and waiting for jobs :3")
}
