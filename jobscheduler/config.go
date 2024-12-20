package jobscheduler

import (
	"fmt"
	"time"
)

// Config contains configuration options for the scheduler
type Config struct {
	// File path for processing log
	ProcessingLogPath string

	// Default number of workers per channel if not specified
	DefaultWorkers int

	// Default timeout for job processing
	DefaultTimeout time.Duration

	// Maximum number of jobs that can be queued per channel
	MaxQueueSize int

	// Working directory for job execution
	WorkDir string

	// Maximum output size to capture from job execution (bytes)
	MaxOutputSize int64

	// Grace period for shutdown
	ShutdownTimeout time.Duration

	// Channel buffer size
	ChannelBufferSize int
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() Config {
	return Config{
		ProcessingLogPath: "processing.log",
		DefaultWorkers:    1,
		DefaultTimeout:    5 * time.Minute,
		MaxQueueSize:      10000,
		WorkDir:           "/tmp/jobscheduler",
		MaxOutputSize:     1024 * 1024, // 1MB
		ShutdownTimeout:   30 * time.Second,
		ChannelBufferSize: 1000,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ProcessingLogPath == "" {
		return fmt.Errorf("processing log path cannot be empty")
	}
	if c.DefaultWorkers < 1 {
		return fmt.Errorf("default workers must be at least 1")
	}
	if c.DefaultTimeout < time.Second {
		return fmt.Errorf("default timeout must be at least 1 second")
	}
	if c.MaxQueueSize < 1 {
		return fmt.Errorf("max queue size must be at least 1")
	}
	if c.WorkDir == "" {
		return fmt.Errorf("work directory cannot be empty")
	}
	if c.ShutdownTimeout < time.Second {
		return fmt.Errorf("shutdown timeout must be at least 1 second")
	}
	if c.ChannelBufferSize < 1 {
		return fmt.Errorf("channel buffer size must be at least 1")
	}
	return nil
}
