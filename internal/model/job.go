package model // import "admin-console/internal/model"
import "time"

type JobType string

const (
	JobTypeRecordLogin JobType = "record_login"
	JobTypeMock        JobType = "mock"
)

type MockJobRequest struct {
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
