package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jonathanleahy/project/jobscheduler"
	"github.com/jonathanleahy/project/webserver/internal/api"
)

// JobsHandler handles job-related requests
type JobsHandler struct {
	scheduler *jobscheduler.Scheduler
}

// NewJobsHandler creates a new jobs handler
func NewJobsHandler(scheduler *jobscheduler.Scheduler) *JobsHandler {
	return &JobsHandler{
		scheduler: scheduler,
	}
}

// ServeHTTP handles HTTP requests for jobs
func (h *JobsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleSubmitJob(w, r)
	case http.MethodGet:
		if strings.Contains(r.URL.Path, "/status/") {
			h.handleJobStatus(w, r)
		} else {
			h.handleListJobs(w, r)
		}
	case http.MethodDelete:
		h.handleCancelJob(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSubmitJob processes job submission requests
func (h *JobsHandler) handleSubmitJob(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req api.SubmitJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Convert API request to scheduler job
	job := jobscheduler.JobPayload{
		ID:      req.JobID,
		Channel: req.Channel,
		Workers: req.Workers,
		Timeout: time.Duration(req.TimeoutSeconds) * time.Second,
		Body:    req.Payload,
	}

	// Add application config if present
	if req.Application != nil {
		job.Application = &jobscheduler.ApplicationConfig{
			Name:        req.Application.Name,
			Path:        req.Application.Path,
			Args:        req.Application.Args,
			Env:         req.Application.Env,
			WorkingDir:  req.Application.WorkingDir,
			PassPayload: req.Application.PassPayload,
		}
	}

	// Submit job
	if err := h.scheduler.SubmitJob(job); err != nil {
		http.Error(w, fmt.Sprintf("Failed to submit job: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := api.SubmitJobResponse{
		JobID:     job.ID,
		Channel:   job.Channel,
		Status:    "accepted",
		Submitted: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// handleJobStatus retrieves the status of a specific job
func (h *JobsHandler) handleJobStatus(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}
	jobID := parts[len(parts)-1]

	// Get job status from scheduler
	status, err := h.scheduler.GetJobStatus(jobID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get job status: %v", err), http.StatusNotFound)
		return
	}

	// Convert to API response
	response := api.JobStatusResponse{
		JobID:     status.ID,
		Channel:   status.Channel,
		Status:    string(status.Status),
		StartTime: status.StartTime,
		EndTime:   status.EndTime,
		Error:     status.Error,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListJobs retrieves a list of jobs
func (h *JobsHandler) handleListJobs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering
	channel := r.URL.Query().Get("channel")
	status := r.URL.Query().Get("status")

	// Get jobs from scheduler
	jobs, err := h.scheduler.ListJobs(channel, status)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list jobs: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to API response
	response := api.ListJobsResponse{
		Jobs: make([]api.JobStatusResponse, len(jobs)),
	}

	for i, job := range jobs {
		response.Jobs[i] = api.JobStatusResponse{
			JobID:     job.ID,
			Channel:   job.Channel,
			Status:    string(job.Status),
			StartTime: job.StartTime,
			EndTime:   job.EndTime,
			Error:     job.Error,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCancelJob cancels a specific job
func (h *JobsHandler) handleCancelJob(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}
	jobID := parts[len(parts)-1]

	// Cancel job
	if err := h.scheduler.CancelJob(jobID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to cancel job: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
