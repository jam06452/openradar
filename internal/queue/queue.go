package queue

import "openradar/internal/domain"

var JobQueue chan *domain.ScanJob

func NewInMemoryQueue(queueSize int) {
	JobQueue = make(chan *domain.ScanJob, queueSize)
}

func Enqueue(job *domain.ScanJob) {
	JobQueue <- job
}

func Dequeue() *domain.ScanJob {
	return <-JobQueue
}
