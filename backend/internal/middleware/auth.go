package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// SupabaseUser represents the user information from Supabase JWT
type SupabaseUser struct {
	ID    string `json:"sub"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// SupabaseClaims represents the JWT claims from Supabase
type SupabaseClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Role  string `json:"role"`
}

// AuthMiddleware validates Supabase JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "MISSING_AUTH_HEADER",
					"message": "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_AUTH_FORMAT",
					"message": "Authorization header must be in format 'Bearer <token>'",
				},
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]
		user, err := validateSupabaseJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": fmt.Sprintf("Invalid or expired token: %v", err),
				},
			})
			c.Abort()
			return
		}

		// Store user information in context for use in handlers
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("userEmail", user.Email)
		
		c.Next()
	}
}

// validateSupabaseJWT validates a Supabase JWT token
func validateSupabaseJWT(tokenString string) (*SupabaseUser, error) {
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET environment variable not set")
	}

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Validate the token and extract claims
	if claims, ok := token.Claims.(*SupabaseClaims); ok && token.Valid {
		// Check if token is expired
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, fmt.Errorf("token has expired")
		}

		user := &SupabaseUser{
			ID:    claims.Subject,
			Email: claims.Email,
			Role:  claims.Role,
		}

		return user, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't require authentication
// It sets user context if a valid token is provided, but allows requests without tokens
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without user context
			c.Next()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, continue without user context
			c.Next()
			return
		}

		tokenString := tokenParts[1]
		user, err := validateSupabaseJWT(tokenString)
		if err != nil {
			// Invalid token, continue without user context
			c.Next()
			return
		}

		// Store user information in context
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("userEmail", user.Email)
		
		c.Next()
	}
}

// GetUserFromContext extracts user information from Gin context
func GetUserFromContext(c *gin.Context) (*SupabaseUser, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	
	supabaseUser, ok := user.(*SupabaseUser)
	return supabaseUser, ok
}

// GetUserIDFromContext extracts user ID from Gin context
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	
	id, ok := userID.(string)
	return id, ok
}

// RequireAuth is a helper function to check if user is authenticated in handlers
func RequireAuth(c *gin.Context) (*SupabaseUser, bool) {
	user, exists := GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "AUTHENTICATION_REQUIRED",
				"message": "Authentication is required for this endpoint",
			},
		})
		return nil, false
	}
	return user, true
}