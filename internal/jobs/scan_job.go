package jobs

import (
	"log"
	"openradar/internal/scanner"
	"time"
)

func scanJobFunc(jobContext JobContext) {
	log.Println("scanning for latest repo updates")
	if _, err := scanner.ScanJob(jobContext.Ctx, jobContext.Cfg.GitHub.Key); err != nil {
		log.Printf("failed to scan for jobs: %v", err)
	}
}

func init() {
	RegisterJob(Job{
		Name:     "Scan for new repositories (/live)",
		Func:     scanJobFunc,
		Schedule: 35 * time.Second,
	})
}
