package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"collaborative-bucket-list/internal/logger"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Checks    map[string]Check  `json:"checks"`
}

// Check represents an individual health check
type Check struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// Health performs a comprehensive health check
func (h *HealthHandler) Health(c *gin.Context) {
	start := time.Now()
	
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   getVersion(),
		Checks:    make(map[string]Check),
	}

	// Database health check
	dbCheck := h.checkDatabase()
	response.Checks["database"] = dbCheck
	
	// Environment health check
	envCheck := h.checkEnvironment()
	response.Checks["environment"] = envCheck

	// Determine overall status
	overallStatus := "healthy"
	for _, check := range response.Checks {
		if check.Status != "healthy" {
			overallStatus = "unhealthy"
			break
		}
	}
	response.Status = overallStatus

	// Log health check
	latency := time.Since(start)
	fields := map[string]interface{}{
		"status":  overallStatus,
		"latency": latency.String(),
	}
	
	if overallStatus == "healthy" {
		logger.Info("Health check completed", fields)
		c.JSON(http.StatusOK, response)
	} else {
		logger.Warn("Health check failed", fields)
		c.JSON(http.StatusServiceUnavailable, response)
	}
}

// Readiness performs a readiness check (lighter than health check)
func (h *HealthHandler) Readiness(c *gin.Context) {
	// Quick database ping
	start := time.Now()
	err := h.db.Ping()
	latency := time.Since(start)
	
	if err != nil {
		logger.Error("Readiness check failed", map[string]interface{}{
			"error":   err.Error(),
			"latency": latency.String(),
		})
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  "database not available",
		})
		return
	}

	logger.Debug("Readiness check passed", map[string]interface{}{
		"latency": latency.String(),
	})
	
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// Liveness performs a liveness check (minimal check)
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// checkDatabase verifies database connectivity
func (h *HealthHandler) checkDatabase() Check {
	start := time.Now()
	
	if h.db == nil {
		return Check{
			Status:  "unhealthy",
			Message: "database connection not initialized",
			Latency: time.Since(start).String(),
		}
	}

	err := h.db.Ping()
	latency := time.Since(start)
	
	if err != nil {
		return Check{
			Status:  "unhealthy",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	// Check if we can execute a simple query
	var result int
	err = h.db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return Check{
			Status:  "unhealthy",
			Message: "database query failed: " + err.Error(),
			Latency: time.Since(start).String(),
		}
	}

	return Check{
		Status:  "healthy",
		Message: "database connection successful",
		Latency: time.Since(start).String(),
	}
}

// checkEnvironment verifies required environment variables
func (h *HealthHandler) checkEnvironment() Check {
	start := time.Now()
	
	requiredVars := []string{
		"DATABASE_URL",
		"SUPABASE_URL",
		"SUPABASE_SERVICE_ROLE_KEY",
		"SUPABASE_JWT_SECRET",
	}

	var missingVars []string
	for _, varName := range requiredVars {
		if os.Getenv(varName) == "" {
			missingVars = append(missingVars, varName)
		}
	}

	if len(missingVars) > 0 {
		return Check{
			Status:  "unhealthy",
			Message: "missing required environment variables: " + joinStrings(missingVars, ", "),
			Latency: time.Since(start).String(),
		}
	}

	return Check{
		Status:  "healthy",
		Message: "all required environment variables present",
		Latency: time.Since(start).String(),
	}
}

// getVersion returns the application version
func getVersion() string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		return "unknown"
	}
	return version
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}