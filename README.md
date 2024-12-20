# Job Scheduler System

A distributed job scheduling system that manages the execution of jobs across multiple channels with support for external application execution and configurable concurrency.

## Features

- Channel-based job processing with configurable workers
- External application execution with environment and working directory support
- JSON payload support for job configuration
- Configurable timeouts and concurrency per channel
- Real-time job status monitoring
- Graceful shutdown handling
- Comprehensive logging
- REST API interface

## Project Structure

```
project/
├── jobscheduler/                  # Job scheduler package
│   ├── cmd/
│   │   └── example/
│   │       └── main.go           # Example implementation
│   ├── internal/
│   │   └── executor/
│   │       └── executor.go       # Application execution
│   ├── config.go                 # Configuration management
│   ├── scheduler.go              # Core scheduler implementation
│   ├── types.go                  # Type definitions
│   ├── processor.go              # Job processing logic
│   └── scheduler_test.go         # Test suite
│
└── webserver/                     # Web server package
    ├── cmd/
    │   └── server/
    │       └── main.go           # Server entry point
    ├── internal/
    │   ├── handlers/            # HTTP handlers
    │   ├── middleware/          # HTTP middleware
    │   └── api/                 # API types
    └── static/                  # Static web files
```

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/project.git

# Build the scheduler
cd project/jobscheduler
go build ./cmd/example

# Build the webserver
cd ../webserver
go build ./cmd/server
```

## Configuration

The job scheduler can be configured through a `Config` struct:

```go
type Config struct {
    ProcessingLogPath string        // Path for processing logs
    DefaultWorkers    int           // Default workers per channel
    DefaultTimeout    time.Duration // Default job timeout
    MaxQueueSize      int          // Maximum queue size per channel
    WorkDir          string        // Working directory for job execution
    MaxOutputSize    int64         // Maximum output size to capture
    ShutdownTimeout  time.Duration // Grace period for shutdown
}
```

## Usage

### Basic Job Submission

```go
scheduler, err := jobscheduler.NewScheduler(jobscheduler.DefaultConfig())
if err != nil {
    log.Fatal(err)
}
defer scheduler.Shutdown()

// Submit a regular job
job := jobscheduler.JobPayload{
    ID:      "job-1",
    Channel: "notifications",
    Body:    json.RawMessage(`{"type": "email", "recipient": "user@example.com"}`),
}

err = scheduler.SubmitJob(job)
```

### External Application Execution

```go
// Submit a job that executes an external application
job := jobscheduler.JobPayload{
    ID:      "process-data",
    Channel: "data-processing",
    Workers: 3,
    Timeout: 10 * time.Minute,
    Application: &jobscheduler.ApplicationConfig{
        Name:       "data-processor",
        Path:       "/usr/local/bin/process-data",
        Args:       []string{"--format", "json"},
        WorkingDir: "/tmp/processing",
        Env: map[string]string{
            "PROCESSING_MODE": "fast",
        },
        PassPayload: true,
    },
    Body: json.RawMessage(`{"input": "data.csv"}`),
}
```

### Channel Statistics

```go
// Get statistics for all channels
stats := scheduler.GetChannelStats()
for channel, stat := range stats {
    fmt.Printf("Channel: %s\n", channel)
    fmt.Printf("Active Jobs: %d\n", len(stat.ActiveJobs))
    fmt.Printf("Total Jobs: %d\n", stat.TotalJobs)
    fmt.Printf("Failed Jobs: %d\n", stat.FailedJobs)
}
```

## API Endpoints

The webserver provides the following REST API endpoints:

### Submit Job
```
POST /api/v1/jobs
Content-Type: application/json

{
    "id": "job-1",
    "channel": "processing",
    "workers": 3,
    "timeout": "5m",
    "application": {
        "name": "processor",
        "path": "/usr/bin/processor",
        "args": ["--mode", "fast"],
        "env": {
            "MODE": "production"
        }
    },
    "body": {
        "data": "example"
    }
}
```

### Get Channel Statistics
```
GET /api/v1/stats
```

### Get Job Status
```
GET /api/v1/jobs/{jobID}
```

## Testing

Run the test suite:

```bash
go test ./... -v
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

[MIT License](LICENSE)# project
