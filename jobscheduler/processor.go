package jobscheduler

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jonathanleahy/project/jobscheduler/internal/executer"
)

// ProcessorConfig contains configuration for a job processor
type ProcessorConfig struct {
	Channel       *Channel
	Executor      *executor.Executor
	ProcessLog    *os.File
	MaxOutputSize int64
}

// Processor handles the processing of jobs for a specific channel
type Processor struct {
	config     ProcessorConfig
	workerPool chan struct{}
	activeJobs sync.Map
}

// NewProcessor creates a new processor instance
func NewProcessor(cfg ProcessorConfig) *Processor {
	return &Processor{
		config:     cfg,
		workerPool: make(chan struct{}, cfg.Channel.Workers),
	}
}

// Start begins processing jobs from the channel
func (p *Processor) Start(ctx context.Context) {
	log.Printf("Starting processor for channel: %s with %d workers",
		p.config.Channel.Name, p.config.Channel.Workers)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Stopping processor for channel: %s", p.config.Channel.Name)
			return
		case job := <-p.config.Channel.Jobs:
			// Wait for available worker
			p.workerPool <- struct{}{}

			// Process job in goroutine
			go func(job JobPayload) {
				defer func() { <-p.workerPool }()
				p.processJob(ctx, job)
			}(job)
		}
	}
}

// processJob handles the execution of a single job
func (p *Processor) processJob(ctx context.Context, job JobPayload) {
	// Store job in active jobs
	p.activeJobs.Store(job.ID, job)
	defer p.activeJobs.Delete(job.ID)

	// Create job context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, p.config.Channel.Timeout)
	defer cancel()

	// Update job status
	job.Status = JobStatusRunning
	job.StartTime = time.Now()

	// Log job start
	p.logJobEvent(job, "STARTED")

	var _ *executor.ExecutionResult
	var err error

	if job.Application != nil {
		// Execute external application
		_, err = p.executeApplication(jobCtx, job)
	} else {
		// Process regular job
		_, err = p.processRegularJob(jobCtx, job)
	}

	// Update job status based on result
	job.EndTime = time.Now()
	if err != nil {
		if err == context.DeadlineExceeded {
			job.Status = JobStatusTimedOut
			job.Error = fmt.Sprintf("job timed out after %v", p.config.Channel.Timeout)
		} else {
			job.Status = JobStatusFailed
			job.Error = err.Error()
		}
	} else {
		job.Status = JobStatusComplete
	}

	// Log job completion
	p.logJobEvent(job, fmt.Sprintf("COMPLETED - Status: %s", job.Status))
}

// executeApplication handles execution of external applications
func (p *Processor) executeApplication(ctx context.Context, job JobPayload) (*executor.ExecutionResult, error) {
	// Create executor config
	cfg := executor.Config{
		Path:        job.Application.Path,
		Args:        job.Application.Args,
		WorkingDir:  job.Application.WorkingDir,
		Env:         job.Application.Env,
		OutputLimit: p.config.MaxOutputSize,
	}

	// Set up stdin if payload should be passed
	if job.Application.PassPayload {
		cfg.Stdin = bytes.NewReader(job.Body)
	}

	// Execute the application
	return p.config.Executor.Execute(ctx, cfg)
}

// processRegularJob handles jobs that don't require external application execution
func (p *Processor) processRegularJob(ctx context.Context, job JobPayload) (*executor.ExecutionResult, error) {
	// Simulate processing for regular jobs
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(100 * time.Millisecond):
		// Create a dummy result for regular jobs
		return &executor.ExecutionResult{
			ExitCode:  0,
			StartTime: job.StartTime,
			EndTime:   time.Now(),
		}, nil
	}
}

// logJobEvent logs a job event to the process log
func (p *Processor) logJobEvent(job JobPayload, event string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Fprintf(p.config.ProcessLog, "%s - %s - Channel: %s, JobID: %s\n",
		timestamp, event, p.config.Channel.Name, job.ID)
}

// GetActiveJobs returns a list of currently active jobs
func (p *Processor) GetActiveJobs() []JobPayload {
	var jobs []JobPayload
	p.activeJobs.Range(func(key, value interface{}) bool {
		if job, ok := value.(JobPayload); ok {
			jobs = append(jobs, job)
		}
		return true
	})
	return jobs
}
