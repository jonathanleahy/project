package jobscheduler

import (
	_ "context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "scheduler-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test config
	cfg := Config{
		ProcessingLogPath: filepath.Join(tmpDir, "processing.log"),
		DefaultWorkers:    2,
		DefaultTimeout:    5 * time.Second,
		MaxQueueSize:      100,
		WorkDir:           tmpDir,
		MaxOutputSize:     1024,
		ShutdownTimeout:   5 * time.Second,
		ChannelBufferSize: 10,
	}

	// Create scheduler
	scheduler, err := NewScheduler(cfg)
	require.NoError(t, err)
	defer scheduler.Shutdown()

	t.Run("SubmitRegularJob", func(t *testing.T) {
		// Create test job
		job := JobPayload{
			ID:      "test-job-1",
			Channel: "test-channel",
			Body:    json.RawMessage(`{"test": true}`),
		}

		// Submit job
		err := scheduler.SubmitJob(job)
		assert.NoError(t, err)

		// Check channel stats
		time.Sleep(100 * time.Millisecond) // Wait for processing
		stats := scheduler.GetChannelStats()
		assert.Contains(t, stats, "test-channel")
		assert.Equal(t, int64(1), stats["test-channel"].TotalJobs)
	})

	t.Run("SubmitApplicationJob", func(t *testing.T) {
		// Create test application job
		job := JobPayload{
			ID:      "test-job-2",
			Channel: "app-channel",
			Workers: 1,
			Timeout: 2 * time.Second,
			Application: &ApplicationConfig{
				Name: "echo",
				Path: "echo",
				Args: []string{"test"},
			},
			Body: json.RawMessage(`{"test": true}`),
		}

		// Submit job
		err := scheduler.SubmitJob(job)
		assert.NoError(t, err)

		// Check channel stats
		time.Sleep(100 * time.Millisecond) // Wait for processing
		stats := scheduler.GetChannelStats()
		assert.Contains(t, stats, "app-channel")
	})

	t.Run("JobTimeout", func(t *testing.T) {
		// Create job with short timeout
		job := JobPayload{
			ID:      "test-job-3",
			Channel: "timeout-channel",
			Timeout: 100 * time.Millisecond,
			Application: &ApplicationConfig{
				Name: "sleep",
				Path: "sleep",
				Args: []string{"1"}, // Sleep for 1 second
			},
		}

		// Submit job
		err := scheduler.SubmitJob(job)
		assert.NoError(t, err)

		// Wait for timeout
		time.Sleep(200 * time.Millisecond)
		stats := scheduler.GetChannelStats()
		assert.Contains(t, stats, "timeout-channel")
	})

	t.Run("ChannelConcurrency", func(t *testing.T) {
		// Create multiple jobs
		for i := 0; i < 5; i++ {
			job := JobPayload{
				ID:      fmt.Sprintf("concurrent-job-%d", i),
				Channel: "concurrent-channel",
				Body:    json.RawMessage(`{"test": true}`),
			}
			err := scheduler.SubmitJob(job)
			assert.NoError(t, err)
		}

		// Check concurrent processing
		time.Sleep(200 * time.Millisecond)
		stats := scheduler.GetChannelStats()
		assert.Contains(t, stats, "concurrent-channel")
		assert.Equal(t, int64(5), stats["concurrent-channel"].TotalJobs)
	})

	t.Run("InvalidJob", func(t *testing.T) {
		// Test job with missing required fields
		job := JobPayload{
			Body: json.RawMessage(`{"test": true}`),
		}

		// Submit should fail
		err := scheduler.SubmitJob(job)
		assert.Error(t, err)
	})

	t.Run("Shutdown", func(t *testing.T) {
		// Submit a job right before shutdown
		job := JobPayload{
			ID:      "shutdown-test",
			Channel: "shutdown-channel",
			Body:    json.RawMessage(`{"test": true}`),
		}
		err := scheduler.SubmitJob(job)
		assert.NoError(t, err)

		// Immediate shutdown
		err = scheduler.Shutdown()
		assert.NoError(t, err)
	})
}

func TestSchedulerConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := DefaultConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("InvalidConfig", func(t *testing.T) {
		cfg := Config{} // Empty config
		assert.Error(t, cfg.Validate())
	})
}
