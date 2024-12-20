// Package executor handles the execution of external applications and processes
package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	_ "strings"
	"sync"
	"time"
)

// ExecutionResult contains the output and status of an executed command
type ExecutionResult struct {
	ExitCode    int
	Stdout      string
	Stderr      string
	StartTime   time.Time
	EndTime     time.Time
	ExecutionID string
}

// Config contains the configuration for executing an application
type Config struct {
	// Application path and arguments
	Path string
	Args []string

	// Execution environment
	WorkingDir string
	Env        map[string]string

	// Input/Output configuration
	Stdin       io.Reader
	OutputLimit int64 // Maximum bytes to capture from stdout/stderr (0 for unlimited)

	// Process management
	KillTimeout time.Duration // Time to wait after sending SIGTERM before SIGKILL
}

// Executor manages the execution of external applications
type Executor struct {
	workDir     string
	mu          sync.RWMutex
	processes   map[string]*exec.Cmd
	execCounter uint64
}

// NewExecutor creates a new executor instance
func NewExecutor(workDir string) (*Executor, error) {
	// Ensure working directory exists and is accessible
	if workDir != "" {
		if err := os.MkdirAll(workDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create working directory: %v", err)
		}
	}

	return &Executor{
		workDir:   workDir,
		processes: make(map[string]*exec.Cmd),
	}, nil
}

// Execute runs an application with the given configuration
func (e *Executor) Execute(ctx context.Context, cfg Config) (*ExecutionResult, error) {
	// Generate unique execution ID
	execID := fmt.Sprintf("exec_%d_%d", time.Now().UnixNano(), e.execCounter)

	// Create command with context
	cmd := exec.CommandContext(ctx, cfg.Path, cfg.Args...)

	// Set up working directory
	if cfg.WorkingDir != "" {
		if !filepath.IsAbs(cfg.WorkingDir) {
			cfg.WorkingDir = filepath.Join(e.workDir, cfg.WorkingDir)
		}
		cmd.Dir = cfg.WorkingDir
	}

	// Set up environment
	cmd.Env = os.Environ() // Start with current environment
	for k, v := range cfg.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set up input if provided
	if cfg.Stdin != nil {
		cmd.Stdin = cfg.Stdin
	}

	// Set up output capture
	var stdout, stderr bytes.Buffer
	if cfg.OutputLimit > 0 {
		cmd.Stdout = &LimitedWriter{W: &stdout, N: cfg.OutputLimit}
		cmd.Stderr = &LimitedWriter{W: &stderr, N: cfg.OutputLimit}
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	// Track the process
	e.mu.Lock()
	e.processes[execID] = cmd
	e.mu.Unlock()

	// Ensure process is removed from tracking when done
	defer func() {
		e.mu.Lock()
		delete(e.processes, execID)
		e.mu.Unlock()
	}()

	// Start execution and track timing
	result := &ExecutionResult{
		ExecutionID: execID,
		StartTime:   time.Now(),
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return result, fmt.Errorf("failed to start process: %v", err)
	}

	// Create a channel for the command completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Wait for either completion or context cancellation
	var execErr error
	select {
	case err := <-done:
		execErr = err
	case <-ctx.Done():
		execErr = e.handleTimeout(cmd, cfg.KillTimeout)
	}

	// Record end time
	result.EndTime = time.Now()

	// Capture output
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// Get exit code
	if execErr != nil {
		if exitErr, ok := execErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1 // Indicate non-exit error
		}
		return result, fmt.Errorf("execution failed: %v", execErr)
	}

	result.ExitCode = 0
	return result, nil
}

// handleTimeout handles graceful shutdown of a process
func (e *Executor) handleTimeout(cmd *exec.Cmd, killTimeout time.Duration) error {
	if cmd.Process == nil {
		return nil
	}

	// Try graceful shutdown first
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		return e.forceKill(cmd)
	}

	// Wait for the process to exit gracefully
	timer := time.NewTimer(killTimeout)
	defer timer.Stop()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-timer.C:
		return e.forceKill(cmd)
	}
}

// forceKill forcefully terminates a process
func (e *Executor) forceKill(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}

// ListProcesses returns information about all running processes
func (e *Executor) ListProcesses() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	processes := make([]string, 0, len(e.processes))
	for execID := range e.processes {
		processes = append(processes, execID)
	}
	return processes
}

// KillProcess terminates a specific process by its execution ID
func (e *Executor) KillProcess(execID string) error {
	e.mu.Lock()
	cmd, exists := e.processes[execID]
	e.mu.Unlock()

	if !exists {
		return fmt.Errorf("process %s not found", execID)
	}

	return e.forceKill(cmd)
}

// Cleanup terminates all running processes
func (e *Executor) Cleanup() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for execID, cmd := range e.processes {
		if err := e.forceKill(cmd); err != nil {
			// Just log the error and continue with cleanup
			fmt.Printf("Failed to kill process %s: %v\n", execID, err)
		}
		delete(e.processes, execID)
	}
}

// LimitedWriter wraps an io.Writer with a byte limit
type LimitedWriter struct {
	W io.Writer // Underlying writer
	N int64     // Max bytes remaining
}

func (l *LimitedWriter) Write(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.ErrShortWrite
	}
	if int64(len(p)) > l.N {
		p = p[:l.N]
	}
	n, err = l.W.Write(p)
	l.N -= int64(n)
	return
}
