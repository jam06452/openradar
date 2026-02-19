package jobs

import (
	"context"
	"openradar/internal/config"
	"time"

	"gorm.io/gorm"
)

type JobContext struct {
	DB  *gorm.DB
	Cfg config.Config
	Ctx context.Context
}

type JobFunc func(jobContext JobContext)

type Job struct {
	Name     string
	Func     JobFunc
	Schedule time.Duration
}

var AllJobs []Job

func RegisterJob(job Job) {
	AllJobs = append(AllJobs, job)
}

func RunJobs(jobContext JobContext) {
	for _, job := range AllJobs {
		go func(j Job) {
			ticker := time.NewTicker(j.Schedule)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					j.Func(jobContext)
				case <-jobContext.Ctx.Done():
					return
				}
			}
		}(job)
	}
}
