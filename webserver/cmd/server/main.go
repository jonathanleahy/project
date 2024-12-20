package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonathanleahy/project/jobscheduler"
	"github.com/jonathanleahy/project/webserver/config"
	"github.com/jonathanleahy/project/webserver/internal/handlers"
	"github.com/jonathanleahy/project/webserver/internal/middleware"
)

var (
	configPath = flag.String("config", "config.yaml", "Path to configuration file")
	port       = flag.Int("port", 8080, "Server port")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize job scheduler
	scheduler, err := jobscheduler.NewScheduler(jobscheduler.Config{
		ProcessingLogPath: cfg.Scheduler.LogPath,
		DefaultWorkers:    cfg.Scheduler.DefaultWorkers,
		DefaultTimeout:    cfg.Scheduler.DefaultTimeout,
		MaxQueueSize:      cfg.Scheduler.MaxQueueSize,
		WorkDir:           cfg.Scheduler.WorkDir,
		MaxOutputSize:     cfg.Scheduler.MaxOutputSize,
		ShutdownTimeout:   cfg.Scheduler.ShutdownTimeout,
	})
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}
	defer scheduler.Shutdown()

	// Create router and handlers
	router := http.NewServeMux()

	// API handlers
	apiHandler := handlers.NewAPIHandler(scheduler)

	// Register routes with middleware
	router.Handle("/api/v1/jobs", middleware.Chain(
		apiHandler.JobsHandler(),
		middleware.Logger,
		middleware.CORS(cfg.Server.AllowedOrigins),
		middleware.Auth(cfg.Server.APIKey),
	))

	router.Handle("/api/v1/stats", middleware.Chain(
		apiHandler.StatsHandler(),
		middleware.Logger,
		middleware.CORS(cfg.Server.AllowedOrigins),
		middleware.Auth(cfg.Server.APIKey),
	))

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	router.Handle("/", fs)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %d", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-stop
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Shutdown scheduler
	if err := scheduler.Shutdown(); err != nil {
		log.Printf("Scheduler shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// healthCheck handles health check requests
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "ok", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
}

// metricsHandler handles metrics requests
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	// TODO: Implement metrics collection and reporting
	fmt.Fprintf(w, "# HELP job_scheduler_active_jobs Number of currently active jobs\n")
	fmt.Fprintf(w, "# TYPE job_scheduler_active_jobs gauge\n")
	// Add actual metrics implementation
}

// debugHandler provides debug information (only in non-production environments)
func debugHandler(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("ENVIRONMENT") == "production" {
		http.Error(w, "Not available in production", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// TODO: Implement debug information
	fmt.Fprintf(w, `{"debug": true, "time": "%s"}`, time.Now().Format(time.RFC3339))
}
