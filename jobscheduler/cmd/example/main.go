package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonathanleahy/project/jobscheduler"
)

var (
	processingLog = flag.String("log", "processing.log", "Path to processing log file")
	workers       = flag.Int("workers", 2, "Default number of workers per channel")
	timeout       = flag.Duration("timeout", 5*time.Minute, "Default job timeout")
)

func main() {
	flag.Parse()

	// Initialize the scheduler
	scheduler, err := jobscheduler.NewScheduler(jobscheduler.Config{
		ProcessingLogPath: *processingLog,
		DefaultWorkers:    *workers,
		DefaultTimeout:    *timeout,
	})
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}
	defer scheduler.Close()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Example jobs channel for demonstration
	jobsChan := make(chan struct{})

	// Start job submission in a separate goroutine
	go submitExampleJobs(scheduler, jobsChan)

	// Wait for shutdown signal
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		close(jobsChan)
	}

	log.Println("Shutting down gracefully...")
}

func submitExampleJobs(scheduler *jobscheduler.Scheduler, done chan struct{}) {
	// Example 1: Simple script execution
	simpleJob := jobscheduler.JobPayload{
		ID:      "simple-job-1",
		Channel: "scripts",
		Workers: 2,
		Timeout: 1 * time.Minute,
		Application: &jobscheduler.ApplicationConfig{
			Name:        "echo",
			Path:        "/bin/echo",
			Args:        []string{"Hello, World!"},
			PassPayload: false,
		},
	}

	err := scheduler.SubmitJob(simpleJob)
	if err != nil {
		log.Printf("Failed to submit simple job: %v", err)
	}

	// Example 2: Data processing job with stdin payload
	processingJob := jobscheduler.JobPayload{
		ID:      "process-data-1",
		Channel: "data-processing",
		Workers: 3,
		Timeout: 5 * time.Minute,
		Application: &jobscheduler.ApplicationConfig{
			Name:       "data-processor",
			Path:       "/usr/local/bin/process-data",
			Args:       []string{"--format", "json", "--compress"},
			WorkingDir: "/tmp/processing",
			Env: map[string]string{
				"PROCESSING_MODE": "fast",
				"MAX_MEMORY":      "1G",
			},
			PassPayload: true,
		},
		Body: json.RawMessage(`{
			"input_file": "data.csv",
			"output_format": "json",
			"compression": true
		}`),
	}

	err = scheduler.SubmitJob(processingJob)
	if err != nil {
		log.Printf("Failed to submit processing job: %v", err)
	}

	// Example 3: Regular job without application execution
	regularJob := jobscheduler.JobPayload{
		ID:      "notification-1",
		Channel: "notifications",
		Body: json.RawMessage(`{
			"type": "email",
			"recipient": "user@example.com",
			"template": "welcome"
		}`),
	}

	err = scheduler.SubmitJob(regularJob)
	if err != nil {
		log.Printf("Failed to submit regular job: %v", err)
	}

	// Example 4: Long-running background job
	backgroundJob := jobscheduler.JobPayload{
		ID:      "backup-1",
		Channel: "backups",
		Workers: 1,
		Timeout: 30 * time.Minute,
		Application: &jobscheduler.ApplicationConfig{
			Name: "backup-script",
			Path: "/usr/local/bin/backup.sh",
			Args: []string{"--full", "--compress"},
			Env: map[string]string{
				"BACKUP_DIR": "/mnt/backups",
			},
		},
	}

	err = scheduler.SubmitJob(backgroundJob)
	if err != nil {
		log.Printf("Failed to submit background job: %v", err)
	}

	// Periodically check and print channel statistics
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := scheduler.GetChannelStats()
			log.Printf("Channel Statistics: %+v", stats)
		case <-done:
			return
		}
	}
}
