package jobscheduler

import (
	"encoding/json"
	"fmt"
	"time"
)

// JobStatus represents the current state of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusComplete  JobStatus = "complete"
	JobStatusFailed    JobStatus = "failed"
	JobStatusTimedOut  JobStatus = "timed_out"
	JobStatusCancelled JobStatus = "cancelled"
)

// JobPayload represents the structure of a job submission
type JobPayload struct {
	ID          string             `json:"id"`
	Channel     string             `json:"channel"`
	Workers     int                `json:"workers,omitempty"`     // Only used for first job in channel
	Timeout     time.Duration      `json:"timeout,omitempty"`     // Only used for first job in channel
	Body        json.RawMessage    `json:"body"`                  // Arbitrary JSON data
	Application *ApplicationConfig `json:"application,omitempty"` // Optional application configuration
	Status      JobStatus          `json:"status"`
	Error       string             `json:"error,omitempty"`
	StartTime   time.Time          `json:"start_time,omitempty"`
	EndTime     time.Time          `json:"end_time,omitempty"`
}

// ApplicationConfig defines the external application to run
type ApplicationConfig struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	PassPayload bool              `json:"pass_payload,omitempty"`
}

// Validate checks if the job payload is valid
func (j *JobPayload) Validate() error {
	if j.ID == "" {
		return fmt.Errorf("job ID cannot be empty")
	}
	if j.Channel == "" {
		return fmt.Errorf("channel cannot be empty")
	}
	if j.Workers < 0 {
		return fmt.Errorf("workers cannot be negative")
	}
	if j.Application != nil {
		if j.Application.Path == "" {
			return fmt.Errorf("application path cannot be empty")
		}
	}
	return nil
}

// ChannelStats represents statistics for a channel
type ChannelStats struct {
	Workers     int       `json:"workers"`
	ActiveJobs  []string  `json:"active_jobs"`
	TotalJobs   int64     `json:"total_jobs"`
	FailedJobs  int64     `json:"failed_jobs"`
	LastJobTime time.Time `json:"last_job_time"`
}

// JobResult represents the result of a job execution
type JobResult struct {
	JobID     string
	Status    JobStatus
	ExitCode  int
	Output    string
	Error     string
	StartTime time.Time
	EndTime   time.Time
}
