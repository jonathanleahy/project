package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete server configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
	Security  SecurityConfig  `yaml:"security"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// ServerConfig contains web server specific configuration
type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
	AllowedOrigins []string      `yaml:"allowed_origins"`
}

// SchedulerConfig contains job scheduler specific configuration
type SchedulerConfig struct {
	LogPath         string        `yaml:"log_path"`
	DefaultWorkers  int           `yaml:"default_workers"`
	DefaultTimeout  time.Duration `yaml:"default_timeout"`
	MaxQueueSize    int           `yaml:"max_queue_size"`
	WorkDir         string        `yaml:"work_dir"`
	MaxOutputSize   int64         `yaml:"max_output_size"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	RetryPolicy     RetryPolicy   `yaml:"retry_policy"`
}

// SecurityConfig contains security related configuration
type SecurityConfig struct {
	APIKey          string          `yaml:"api_key"`
	TokenExpiry     time.Duration   `yaml:"token_expiry"`
	EnableTLS       bool            `yaml:"enable_tls"`
	TLSCert         string          `yaml:"tls_cert"`
	TLSKey          string          `yaml:"tls_key"`
	RateLimit       RateLimitConfig `yaml:"rate_limit"`
	AllowedIPRanges []string        `yaml:"allowed_ip_ranges"`
}

// LoggingConfig contains logging related configuration
type LoggingConfig struct {
	Level        string `yaml:"level"`
	Format       string `yaml:"format"`
	FilePath     string `yaml:"file_path"`
	MaxSize      int    `yaml:"max_size"`    // megabytes
	MaxAge       int    `yaml:"max_age"`     // days
	MaxBackups   int    `yaml:"max_backups"` // files
	Compress     bool   `yaml:"compress"`
	EnableStdout bool   `yaml:"enable_stdout"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool `yaml:"enabled"`
	RequestsPerMin int  `yaml:"requests_per_min"`
	BurstSize      int  `yaml:"burst_size"`
}

// RetryPolicy contains job retry configuration
type RetryPolicy struct {
	MaxRetries    int           `yaml:"max_retries"`
	InitialDelay  time.Duration `yaml:"initial_delay"`
	MaxDelay      time.Duration `yaml:"max_delay"`
	BackoffFactor float64       `yaml:"backoff_factor"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	// Read configuration file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse YAML
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Set defaults for missing values
	config.setDefaults()

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return config, nil
}

// setDefaults sets default values for missing configuration
func (c *Config) setDefaults() {
	// Server defaults
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 15 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 15 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 60 * time.Second
	}
	if c.Server.MaxHeaderBytes == 0 {
		c.Server.MaxHeaderBytes = 1 << 20 // 1MB
	}

	// Scheduler defaults
	if c.Scheduler.DefaultWorkers == 0 {
		c.Scheduler.DefaultWorkers = 1
	}
	if c.Scheduler.DefaultTimeout == 0 {
		c.Scheduler.DefaultTimeout = 5 * time.Minute
	}
	if c.Scheduler.MaxQueueSize == 0 {
		c.Scheduler.MaxQueueSize = 10000
	}
	if c.Scheduler.MaxOutputSize == 0 {
		c.Scheduler.MaxOutputSize = 1 << 20 // 1MB
	}
	if c.Scheduler.ShutdownTimeout == 0 {
		c.Scheduler.ShutdownTimeout = 30 * time.Second
	}

	// Security defaults
	if c.Security.TokenExpiry == 0 {
		c.Security.TokenExpiry = 24 * time.Hour
	}
	if c.Security.RateLimit.RequestsPerMin == 0 {
		c.Security.RateLimit.RequestsPerMin = 60
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.MaxSize == 0 {
		c.Logging.MaxSize = 100
	}
	if c.Logging.MaxAge == 0 {
		c.Logging.MaxAge = 7
	}
	if c.Logging.MaxBackups == 0 {
		c.Logging.MaxBackups = 5
	}
}

// validate checks if the configuration is valid
func (c *Config) validate() error {
	// Validate Server configuration
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Server.Port)
	}

	// Validate Scheduler configuration
	if c.Scheduler.DefaultWorkers < 1 {
		return fmt.Errorf("default workers must be at least 1")
	}
	if c.Scheduler.MaxQueueSize < 1 {
		return fmt.Errorf("max queue size must be at least 1")
	}

	// Validate Security configuration
	if c.Security.EnableTLS {
		if c.Security.TLSCert == "" || c.Security.TLSKey == "" {
			return fmt.Errorf("TLS certificate and key are required when TLS is enabled")
		}
	}
	if c.Security.RateLimit.Enabled && c.Security.RateLimit.RequestsPerMin < 1 {
		return fmt.Errorf("requests per minute must be at least 1")
	}

	return nil
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	config := &Config{}

	// Load from environment variables
	if port := os.Getenv("SERVER_PORT"); port != "" {
		// Parse and set port
	}
	if workers := os.Getenv("SCHEDULER_WORKERS"); workers != "" {
		// Parse and set workers
	}
	// ... add more environment variables as needed

	config.setDefaults()
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}
