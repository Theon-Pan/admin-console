package model // import "admin-console/internal/model"
import "time"

type JobType string

const (
	JobRecordLogin JobType = "record_login"
	JobMock        JobType = "mock"
)

type MockJobArg struct {
	SleepTime     time.Duration
	ReturnSuccess bool
}

// Job represents a payload sent to the processing queue.
type Job struct {
	UserID int64
	Type   JobType
	JobArg *any
}

// JobList represents a list of jobs.
type JobList []Job
