package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/yourusername/project/jobscheduler"
	"github.com/yourusername/project/webserver/internal/api"
)

// StatsHandler handles statistics-related requests
type StatsHandler struct {
	scheduler *jobscheduler.Scheduler
}

// NewStatsHandler creates a new statistics handler
func NewStatsHandler(scheduler *jobscheduler.Scheduler) *StatsHandler {
	return &StatsHandler{
		scheduler: scheduler,
	}
}

// ServeHTTP handles HTTP requests for statistics
func (h *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/channels"):
		h.handleChannelStats(w, r)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/summary"):
		h.handleSummaryStats(w, r)
	case r.Method == http.MethodGet:
		h.handleOverallStats(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleChannelStats returns statistics for specific channels
func (h *StatsHandler) handleChannelStats(w http.ResponseWriter, r *http.Request) {
	// Get channel name from query parameter
	channel := r.URL.Query().Get("channel")

	// Get channel statistics from scheduler
	stats := h.scheduler.GetChannelStats()

	// Filter by channel if specified
	response := make(map[string]api.ChannelStats)
	if channel != "" {
		if channelStats, ok := stats[channel]; ok {
			response[channel] = convertToAPIStats(channelStats)
		}
	} else {
		for ch, stat := range stats {
			response[ch] = convertToAPIStats(stat)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSummaryStats returns summarized statistics
func (h *StatsHandler) handleSummaryStats(w http.ResponseWriter, r *http.Request) {
	// Get time range from query parameters
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	// Get summary statistics from scheduler
	summary, err := h.scheduler.GetStatsSummary(from, to)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get stats summary: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to API response
	response := api.StatsSummary{
		TotalJobs:      summary.TotalJobs,
		CompletedJobs:  summary.CompletedJobs,
		FailedJobs:     summary.FailedJobs,
		AverageRuntime: summary.AverageRuntime,
		ActiveChannels: summary.ActiveChannels,
		TimeRange: api.TimeRange{
			From: from,
			To:   to,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleOverallStats returns overall statistics
func (h *StatsHandler) handleOverallStats(w http.ResponseWriter, r *http.Request) {
	// Get overall statistics from scheduler
	stats := h.scheduler.GetOverallStats()

	// Convert to API response
	response := api.OverallStats{
		ActiveJobs:     stats.ActiveJobs,
		QueuedJobs:     stats.QueuedJobs,
		CompletedJobs:  stats.CompletedJobs,
		FailedJobs:     stats.FailedJobs,
		ActiveChannels: stats.ActiveChannels,
		Uptime:         stats.Uptime.String(),
		LastUpdate:     stats.LastUpdate,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// convertToAPIStats converts internal stats to API format
func convertToAPIStats(stats *jobscheduler.ChannelStats) api.ChannelStats {
	return api.ChannelStats{
		Workers:     stats.Workers,
		ActiveJobs:  stats.ActiveJobs,
		TotalJobs:   stats.TotalJobs,
		FailedJobs:  stats.FailedJobs,
		LastJobTime: stats.LastJobTime,
	}
}
