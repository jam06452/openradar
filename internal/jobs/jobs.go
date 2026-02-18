package jobs

import (
	"time"

	"gorm.io/gorm"
)

type JobFunc func(dbJOB *gorm.DB)

var AllJobs []JobFunc

func executeAllJobs(dbJOB *gorm.DB) {
	for _, job := range AllJobs {
		job(dbJOB)
	}
}

func RunAllJobsEvery30Minutes(dbJOB *gorm.DB) {
	ticker := time.NewTicker(30 * time.Minute) // 30 minutes
	defer ticker.Stop()

	for range ticker.C {
		executeAllJobs(dbJOB)
	}
}
