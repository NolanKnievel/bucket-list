package main

import (
	"collaborative-bucket-list/internal/handlers"
	"collaborative-bucket-list/internal/middleware"
	"collaborative-bucket-list/internal/repositories"
	"collaborative-bucket-list/pkg/database"
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database connection
	dbConfig := database.LoadConfigFromEnv()
	if err := database.Connect(dbConfig); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Run database migrations
	if err := database.RunMigrations("migrations"); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}

	// Initialize repository manager
	repoManager := repositories.NewPostgresRepositoryManager(database.DB)

	// Initialize handlers
	groupHandler := handlers.NewGroupHandler(repoManager)

	// Set up Gin router
	r := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins != "" {
		corsConfig.AllowOrigins = strings.Split(allowedOrigins, ",")
	} else {
		corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(corsConfig))

	// Basic health check endpoint
	r.GET("/health", func(c *gin.Context) {
		// Check database health
		if err := database.HealthCheck(); err != nil {
			c.JSON(503, gin.H{
				"status": "error",
				"message": "Database health check failed",
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"message": "Collaborative Bucket List API is running",
			"database": "connected",
		})
	})

	// Auth verification endpoint
	r.POST("/api/auth/verify", middleware.AuthMiddleware(), func(c *gin.Context) {
		user, exists := middleware.GetUserFromContext(c)
		if !exists {
			c.JSON(500, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to get user from context",
				},
			})
			return
		}

		c.JSON(200, gin.H{
			"user":  user,
			"valid": true,
		})
	})

	// Group management endpoints
	api := r.Group("/api")
	{
		// POST /api/groups - Create new group (requires authentication)
		api.POST("/groups", middleware.AuthMiddleware(), groupHandler.CreateGroup)
		
		// GET /api/groups/:id - Get group details
		api.GET("/groups/:id", groupHandler.GetGroup)
		
		// GET /api/users/groups - Get user's groups (requires authentication)
		api.GET("/users/groups", middleware.AuthMiddleware(), groupHandler.GetUserGroups)
	}

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}