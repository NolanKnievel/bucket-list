package metrics

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"collaborative-bucket-list/internal/logger"

	"github.com/gin-gonic/gin"
)

// Metrics holds application metrics
type Metrics struct {
	mu                sync.RWMutex
	RequestCount      int64             `json:"request_count"`
	ErrorCount        int64             `json:"error_count"`
	ActiveConnections int64             `json:"active_connections"`
	ResponseTimes     []time.Duration   `json:"-"`
	AverageResponse   time.Duration     `json:"average_response_time"`
	Uptime           time.Duration     `json:"uptime"`
	StartTime        time.Time         `json:"start_time"`
	LastRequest      time.Time         `json:"last_request"`
	StatusCodes      map[int]int64     `json:"status_codes"`
	Endpoints        map[string]int64  `json:"endpoints"`
}

var (
	globalMetrics *Metrics
	once          sync.Once
)

// Init initializes the metrics system
func Init() {
	once.Do(func() {
		globalMetrics = &Metrics{
			StartTime:   time.Now(),
			StatusCodes: make(map[int]int64),
			Endpoints:   make(map[string]int64),
		}
		
		// Start background metrics collection if enabled
		if isMetricsEnabled() {
			go collectSystemMetrics()
		}
	})
}

// isMetricsEnabled checks if metrics collection is enabled
func isMetricsEnabled() bool {
	enabled := os.Getenv("METRICS_ENABLED")
	return enabled == "true" || enabled == "1"
}

// IncrementRequests increments the request counter
func IncrementRequests() {
	if globalMetrics == nil {
		return
	}
	
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	
	globalMetrics.RequestCount++
	globalMetrics.LastRequest = time.Now()
}

// IncrementErrors increments the error counter
func IncrementErrors() {
	if globalMetrics == nil {
		return
	}
	
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	
	globalMetrics.ErrorCount++
}

// RecordResponseTime records a response time
func RecordResponseTime(duration time.Duration) {
	if globalMetrics == nil {
		return
	}
	
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	
	// Keep only last 1000 response times for average calculation
	if len(globalMetrics.ResponseTimes) >= 1000 {
		globalMetrics.ResponseTimes = globalMetrics.ResponseTimes[1:]
	}
	
	globalMetrics.ResponseTimes = append(globalMetrics.ResponseTimes, duration)
	
	// Calculate average
	var total time.Duration
	for _, rt := range globalMetrics.ResponseTimes {
		total += rt
	}
	globalMetrics.AverageResponse = total / time.Duration(len(globalMetrics.ResponseTimes))
}

// RecordStatusCode records a status code
func RecordStatusCode(code int) {
	if globalMetrics == nil {
		return
	}
	
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	
	globalMetrics.StatusCodes[code]++
	
	if code >= 400 {
		globalMetrics.ErrorCount++
	}
}

// RecordEndpoint records an endpoint hit
func RecordEndpoint(endpoint string) {
	if globalMetrics == nil {
		return
	}
	
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	
	globalMetrics.Endpoints[endpoint]++
}

// IncrementConnections increments active connections
func IncrementConnections() {
	if globalMetrics == nil {
		return
	}
	
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	
	globalMetrics.ActiveConnections++
}

// DecrementConnections decrements active connections
func DecrementConnections() {
	if globalMetrics == nil {
		return
	}
	
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	
	if globalMetrics.ActiveConnections > 0 {
		globalMetrics.ActiveConnections--
	}
}

// GetMetrics returns current metrics
func GetMetrics() *Metrics {
	if globalMetrics == nil {
		return nil
	}
	
	globalMetrics.mu.RLock()
	defer globalMetrics.mu.RUnlock()
	
	// Create a copy to avoid race conditions
	metrics := &Metrics{
		RequestCount:      globalMetrics.RequestCount,
		ErrorCount:        globalMetrics.ErrorCount,
		ActiveConnections: globalMetrics.ActiveConnections,
		AverageResponse:   globalMetrics.AverageResponse,
		Uptime:           time.Since(globalMetrics.StartTime),
		StartTime:        globalMetrics.StartTime,
		LastRequest:      globalMetrics.LastRequest,
		StatusCodes:      make(map[int]int64),
		Endpoints:        make(map[string]int64),
	}
	
	// Copy maps
	for k, v := range globalMetrics.StatusCodes {
		metrics.StatusCodes[k] = v
	}
	for k, v := range globalMetrics.Endpoints {
		metrics.Endpoints[k] = v
	}
	
	return metrics
}

// MetricsMiddleware returns a Gin middleware for collecting metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isMetricsEnabled() {
			c.Next()
			return
		}
		
		start := time.Now()
		
		IncrementRequests()
		RecordEndpoint(c.Request.Method + " " + c.FullPath())
		
		c.Next()
		
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		
		RecordResponseTime(duration)
		RecordStatusCode(statusCode)
	}
}

// MetricsHandler returns metrics in JSON format
func MetricsHandler(c *gin.Context) {
	if !isMetricsEnabled() {
		c.JSON(http.StatusNotFound, gin.H{"error": "metrics not enabled"})
		return
	}
	
	metrics := GetMetrics()
	if metrics == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "metrics not initialized"})
		return
	}
	
	// Add system metrics
	systemMetrics := getSystemMetrics()
	
	response := gin.H{
		"application": metrics,
		"system":      systemMetrics,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, response)
}

// SystemMetrics holds system-level metrics
type SystemMetrics struct {
	GoVersion      string  `json:"go_version"`
	Goroutines     int     `json:"goroutines"`
	MemoryAlloc    uint64  `json:"memory_alloc_bytes"`
	MemoryTotal    uint64  `json:"memory_total_bytes"`
	MemorySys      uint64  `json:"memory_sys_bytes"`
	GCRuns         uint32  `json:"gc_runs"`
	CPUCount       int     `json:"cpu_count"`
}

// getSystemMetrics returns current system metrics
func getSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return SystemMetrics{
		GoVersion:   runtime.Version(),
		Goroutines:  runtime.NumGoroutine(),
		MemoryAlloc: m.Alloc,
		MemoryTotal: m.TotalAlloc,
		MemorySys:   m.Sys,
		GCRuns:      m.NumGC,
		CPUCount:    runtime.NumCPU(),
	}
}

// collectSystemMetrics runs in background to collect system metrics
func collectSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		metrics := getSystemMetrics()
		
		// Log system metrics periodically
		fields := map[string]interface{}{
			"goroutines":   metrics.Goroutines,
			"memory_alloc": metrics.MemoryAlloc,
			"gc_runs":      metrics.GCRuns,
		}
		
		logger.Debug("System metrics collected", fields)
	}
}

// StartMetricsServer starts a separate HTTP server for metrics
func StartMetricsServer() {
	if !isMetricsEnabled() {
		return
	}
	
	port := os.Getenv("METRICS_PORT")
	if port == "" {
		port = "9090"
	}
	
	// Validate port
	if _, err := strconv.Atoi(port); err != nil {
		logger.Error("Invalid metrics port", map[string]interface{}{
			"port":  port,
			"error": err.Error(),
		})
		return
	}
	
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := GetMetrics()
		if metrics == nil {
			http.Error(w, "metrics not initialized", http.StatusInternalServerError)
			return
		}
		
		systemMetrics := getSystemMetrics()
		
		response := map[string]interface{}{
			"application": metrics,
			"system":      systemMetrics,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	
	logger.Info("Starting metrics server", map[string]interface{}{
		"port": port,
	})
	
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Metrics server failed", map[string]interface{}{
			"error": err.Error(),
		})
	}
}