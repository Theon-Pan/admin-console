package model // import "admin-console/internal/model"

// Job represents a payload sent to the processing queue.
type Job struct {
	UserID int64
}

// JobList represents a list of jobs.
type JobList []Job
