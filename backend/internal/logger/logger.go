package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger represents the application logger
type Logger struct {
	level  LogLevel
	format string
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

var defaultLogger *Logger

// Init initializes the logger with configuration from environment
func Init() {
	level := getLogLevel(os.Getenv("LOG_LEVEL"))
	format := getLogFormat(os.Getenv("LOG_FORMAT"))
	
	defaultLogger = &Logger{
		level:  level,
		format: format,
	}
	
	// Set Gin mode based on log level
	if level <= DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

// getLogLevel converts string to LogLevel
func getLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

// getLogFormat determines log format
func getLogFormat(format string) string {
	switch strings.ToLower(format) {
	case "json":
		return "json"
	default:
		return "text"
	}
}

// Debug logs a debug message
func Debug(message string, fields ...map[string]interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.log(DEBUG, message, fields...)
}

// Info logs an info message
func Info(message string, fields ...map[string]interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.log(INFO, message, fields...)
}

// Warn logs a warning message
func Warn(message string, fields ...map[string]interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.log(WARN, message, fields...)
}

// Error logs an error message
func Error(message string, fields ...map[string]interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.log(ERROR, message, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(message string, fields ...map[string]interface{}) {
	if defaultLogger == nil {
		Init()
	}
	defaultLogger.log(FATAL, message, fields...)
	os.Exit(1)
}

// log performs the actual logging
func (l *Logger) log(level LogLevel, message string, fields ...map[string]interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	levelStr := l.levelToString(level)

	if l.format == "json" {
		entry := LogEntry{
			Timestamp: timestamp,
			Level:     levelStr,
			Message:   message,
		}

		if len(fields) > 0 {
			entry.Fields = fields[0]
		}

		jsonData, err := json.Marshal(entry)
		if err != nil {
			log.Printf("Error marshaling log entry: %v", err)
			return
		}

		fmt.Println(string(jsonData))
	} else {
		// Text format
		logMsg := fmt.Sprintf("[%s] %s: %s", timestamp, levelStr, message)
		
		if len(fields) > 0 {
			for key, value := range fields[0] {
				logMsg += fmt.Sprintf(" %s=%v", key, value)
			}
		}
		
		fmt.Println(logMsg)
	}
}

// levelToString converts LogLevel to string
func (l *Logger) levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// GinLogger returns a Gin middleware for logging HTTP requests
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Build log fields
		fields := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       path,
			"status":     statusCode,
			"latency":    latency.String(),
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}

		if raw != "" {
			fields["query"] = raw
		}

		// Log based on status code
		message := fmt.Sprintf("%s %s", c.Request.Method, path)
		
		if statusCode >= 500 {
			Error(message, fields)
		} else if statusCode >= 400 {
			Warn(message, fields)
		} else {
			Info(message, fields)
		}
	}
}

// Recovery returns a Gin middleware for panic recovery
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fields := map[string]interface{}{
					"error":  fmt.Sprintf("%v", err),
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
					"ip":     c.ClientIP(),
				}
				
				Error("Panic recovered", fields)
				c.JSON(500, gin.H{"error": "Internal server error"})
				c.Abort()
			}
		}()
		c.Next()
	}
}