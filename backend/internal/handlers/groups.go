package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"collaborative-bucket-list/internal/middleware"
	"collaborative-bucket-list/internal/models"
	"collaborative-bucket-list/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GroupHandler handles group-related HTTP requests
type GroupHandler struct {
	repos repositories.RepositoryManager
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(repos repositories.RepositoryManager) *GroupHandler {
	return &GroupHandler{repos: repos}
}

// CreateGroup handles POST /api/groups
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	// Require authentication
	user, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	var req models.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST_BODY",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// Sanitize input
	req.Sanitize()

	// Validate request
	validation := req.Validate()
	if !validation.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Request validation failed",
				"details": validation.Errors,
			},
		})
		return
	}

	// Create group model
	group := &models.Group{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Deadline:  req.Deadline,
		CreatedAt: time.Now(),
		CreatedBy: user.ID,
	}

	// Create group and creator member in a transaction
	err := h.repos.WithTx(c.Request.Context(), func(tx *sql.Tx) error {
		// Create transactional repository manager
		txRepos := repositories.NewTransactionalRepositoryManager(tx)
		
		// Create the group
		if err := txRepos.Groups().Create(c.Request.Context(), group); err != nil {
			return fmt.Errorf("failed to create group: %w", err)
		}

		// Create the creator as the first member
		member := &models.Member{
			ID:        uuid.New().String(),
			GroupID:   group.ID,
			UserID:    &user.ID,
			Name:      user.Email, // Use email as default name for authenticated users
			JoinedAt:  time.Now(),
			IsCreator: true,
		}

		if err := txRepos.Members().Create(c.Request.Context(), member); err != nil {
			return fmt.Errorf("failed to create creator member: %w", err)
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "GROUP_CREATION_FAILED",
				"message": "Failed to create group",
				"details": err.Error(),
			},
		})
		return
	}

	// Generate shareable link
	shareLink := fmt.Sprintf("%s/groups/%s", getFrontendURL(), group.ID)

	// Return created group with share link
	response := gin.H{
		"id":        group.ID,
		"name":      group.Name,
		"deadline":  group.Deadline,
		"createdAt": group.CreatedAt,
		"createdBy": group.CreatedBy,
		"shareLink": shareLink,
	}

	c.JSON(http.StatusCreated, response)
}

// GetGroup handles GET /api/groups/:id
func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_GROUP_ID",
				"message": "Group ID is required",
			},
		})
		return
	}

	// Validate UUID format
	validation := models.ValidateUUID(groupID)
	if !validation.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_GROUP_ID",
				"message": "Invalid group ID format",
				"details": validation.Errors,
			},
		})
		return
	}

	// Get group with details
	groupDetails, err := h.repos.Groups().GetWithDetails(c.Request.Context(), groupID)
	if err != nil {
		if err.Error() == fmt.Sprintf("group not found: %s", groupID) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "GROUP_NOT_FOUND",
					"message": "Group not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "GROUP_RETRIEVAL_FAILED",
				"message": "Failed to retrieve group",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, groupDetails)
}

// GetUserGroups handles GET /api/users/groups
func (h *GroupHandler) GetUserGroups(c *gin.Context) {
	// Require authentication
	user, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	// Get group summaries for the user
	summaries, err := h.repos.Groups().GetSummariesByUserID(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "USER_GROUPS_RETRIEVAL_FAILED",
				"message": "Failed to retrieve user groups",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": summaries,
	})
}

// JoinGroup handles POST /api/groups/:id/join
func (h *GroupHandler) JoinGroup(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_GROUP_ID",
				"message": "Group ID is required",
			},
		})
		return
	}

	// Validate UUID format
	validation := models.ValidateUUID(groupID)
	if !validation.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_GROUP_ID",
				"message": "Invalid group ID format",
				"details": validation.Errors,
			},
		})
		return
	}

	var req models.JoinGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_REQUEST_BODY",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// Sanitize input
	req.Sanitize()

	// Validate request
	validation = req.Validate()
	if !validation.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Request validation failed",
				"details": validation.Errors,
			},
		})
		return
	}

	// Check if group exists
	_, err := h.repos.Groups().GetByID(c.Request.Context(), groupID)
	if err != nil {
		if err.Error() == fmt.Sprintf("group not found: %s", groupID) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "GROUP_NOT_FOUND",
					"message": "Group not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "GROUP_RETRIEVAL_FAILED",
				"message": "Failed to retrieve group",
				"details": err.Error(),
			},
		})
		return
	}

	// Handle authenticated vs anonymous users
	var userID *string
	if req.UserID != nil && *req.UserID != "" {
		userID = req.UserID
		
		// Check if authenticated user is already a member
		exists, err := h.repos.Members().ExistsByGroupAndUser(c.Request.Context(), groupID, *userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "MEMBERSHIP_CHECK_FAILED",
					"message": "Failed to check existing membership",
					"details": err.Error(),
				},
			})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "ALREADY_MEMBER",
					"message": "User is already a member of this group",
				},
			})
			return
		}
	}

	// Create new member
	member := &models.Member{
		ID:        uuid.New().String(),
		GroupID:   groupID,
		UserID:    userID,
		Name:      req.MemberName,
		JoinedAt:  time.Now(),
		IsCreator: false,
	}

	// Create member in database
	if err := h.repos.Members().Create(c.Request.Context(), member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "MEMBER_CREATION_FAILED",
				"message": "Failed to add member to group",
				"details": err.Error(),
			},
		})
		return
	}

	// Return created member
	c.JSON(http.StatusCreated, gin.H{
		"member": member,
	})
}

// getBaseURL extracts the base URL from the request
func getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	
	host := c.Request.Host
	if forwardedHost := c.GetHeader("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}
	
	return fmt.Sprintf("%s://%s", scheme, host)
}

// getFrontendURL returns the frontend URL from environment variable
func getFrontendURL() string {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		// Default to localhost:3000 if not set
		frontendURL = "http://localhost:3000"
	}
	return frontendURL
}