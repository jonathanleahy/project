package api

import (
	"encoding/json"
	"fmt"
	"time"
)

// SubmitJobRequest represents the request structure for job submission
type SubmitJobRequest struct {
	JobID          string             `json:"job_id"`
	Channel        string             `json:"channel"`
	Workers        int                `json:"workers,omitempty"`
	TimeoutSeconds int                `json:"timeout_seconds,omitempty"`
	Priority       int                `json:"priority,omitempty"`
	Application    *ApplicationConfig `json:"application,omitempty"`
	Payload        json.RawMessage    `json:"payload"`
	Tags           []string           `json:"tags,omitempty"`
	Notify         *NotifyConfig      `json:"notify,omitempty"`
}

// ApplicationConfig defines external application configuration
type ApplicationConfig struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty"`
	PassPayload bool              `json:"pass_payload,omitempty"`
	Timeout     int               `json:"timeout,omitempty"`
}

// NotifyConfig defines notification settings for job events
type NotifyConfig struct {
	Webhook  string   `json:"webhook,omitempty"`
	Email    string   `json:"email,omitempty"`
	Events   []string `json:"events,omitempty"` // completed, failed, timeout
	Template string   `json:"template,omitempty"`
}

// SubmitJobResponse represents the response structure for job submission
type SubmitJobResponse struct {
	JobID     string    `json:"job_id"`
	Channel   string    `json:"channel"`
	Status    string    `json:"status"`
	Submitted time.Time `json:"submitted"`
	QueueSize int       `json:"queue_size,omitempty"`
}

// JobStatusResponse represents the response structure for job status
type JobStatusResponse struct {
	JobID      string    `json:"job_id"`
	Channel    string    `json:"channel"`
	Status     string    `json:"status"`
	Progress   float64   `json:"progress,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	Duration   string    `json:"duration,omitempty"`
	Error      string    `json:"error,omitempty"`
	Logs       []string  `json:"logs,omitempty"`
	ExitCode   int       `json:"exit_code,omitempty"`
	RetryCount int       `json:"retry_count,omitempty"`
}

// ListJobsResponse represents the response structure for listing jobs
type ListJobsResponse struct {
	Jobs       []JobStatusResponse `json:"jobs"`
	TotalJobs  int                 `json:"total_jobs"`
	PageSize   int                 `json:"page_size"`
	PageNumber int                 `json:"page_number"`
}

// ChannelStats represents statistics for a channel
type ChannelStats struct {
	Workers     int       `json:"workers"`
	ActiveJobs  []string  `json:"active_jobs"`
	TotalJobs   int64     `json:"total_jobs"`
	FailedJobs  int64     `json:"failed_jobs"`
	LastJobTime time.Time `json:"last_job_time"`
	Uptime      string    `json:"uptime"`
	QueueSize   int       `json:"queue_size"`
	Throughput  float64   `json:"throughput"` // jobs per minute
}

// StatsSummary represents summarized statistics
type StatsSummary struct {
	TotalJobs      int64     `json:"total_jobs"`
	CompletedJobs  int64     `json:"completed_jobs"`
	FailedJobs     int64     `json:"failed_jobs"`
	AverageRuntime float64   `json:"average_runtime"`
	ActiveChannels int       `json:"active_channels"`
	TimeRange      TimeRange `json:"time_range"`
}

// TimeRange represents a time period for statistics
type TimeRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// OverallStats represents system-wide statistics
type OverallStats struct {
	ActiveJobs     int         `json:"active_jobs"`
	QueuedJobs     int         `json:"queued_jobs"`
	CompletedJobs  int64       `json:"completed_jobs"`
	FailedJobs     int64       `json:"failed_jobs"`
	ActiveChannels int         `json:"active_channels"`
	Uptime         string      `json:"uptime"`
	LastUpdate     time.Time   `json:"last_update"`
	SystemStats    SystemStats `json:"system_stats"`
}

// SystemStats represents system resource statistics
type SystemStats struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	GoRoutines  int     `json:"goroutines"`
}

// Validate performs validation on the job submission request
func (r *SubmitJobRequest) Validate() error {
	if r.JobID == "" {
		return fmt.Errorf("job_id is required")
	}
	if r.Channel == "" {
		return fmt.Errorf("channel is required")
	}
	if r.Workers < 0 {
		return fmt.Errorf("workers cannot be negative")
	}
	if r.TimeoutSeconds < 0 {
		return fmt.Errorf("timeout_seconds cannot be negative")
	}
	if r.Application != nil {
		if r.Application.Path == "" {
			return fmt.Errorf("application path is required")
		}
	}
	if r.Priority < 0 || r.Priority > 10 {
		return fmt.Errorf("priority must be between 0 and 10")
	}
	if r.Notify != nil {
		if r.Notify.Webhook == "" && r.Notify.Email == "" {
			return fmt.Errorf("either webhook or email must be specified for notifications")
		}
	}
	return nil
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error      string `json:"error"`
	Code       int    `json:"code"`
	RequestID  string `json:"request_id,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// JobQuery represents query parameters for job listing
type JobQuery struct {
	Channel    string    `json:"channel"`
	Status     string    `json:"status"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Tags       []string  `json:"tags"`
	PageSize   int       `json:"page_size"`
	PageNumber int       `json:"page_number"`
	SortBy     string    `json:"sort_by"`
	SortOrder  string    `json:"sort_order"`
}

// BatchJobRequest represents a request to submit multiple jobs
type BatchJobRequest struct {
	Jobs     []SubmitJobRequest `json:"jobs"`
	Parallel bool               `json:"parallel"`
}

// BatchJobResponse represents the response for a batch job submission
type BatchJobResponse struct {
	JobIDs   []string `json:"job_ids"`
	Accepted int      `json:"accepted"`
	Rejected int      `json:"rejected"`
	Errors   []error  `json:"errors,omitempty"`
	BatchID  string   `json:"batch_id"`
}
