package handlers

import (
	"fmt"
	"net/http"
	"time"

	"collaborative-bucket-list/internal/models"
	"collaborative-bucket-list/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BucketItemHandler handles bucket list item-related HTTP requests
type BucketItemHandler struct {
	repos repositories.RepositoryManager
}

// NewBucketItemHandler creates a new bucket item handler
func NewBucketItemHandler(repos repositories.RepositoryManager) *BucketItemHandler {
	return &BucketItemHandler{repos: repos}
}

// CreateItem handles POST /api/groups/:id/items
func (h *BucketItemHandler) CreateItem(c *gin.Context) {
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

	var req models.CreateItemRequest
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

	// Validate member ID format
	memberValidation := models.ValidateUUID(req.MemberID)
	if !memberValidation.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_MEMBER_ID",
				"message": "Invalid member ID format",
				"details": memberValidation.Errors,
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

	// Check if member exists and belongs to the group
	member, err := h.repos.Members().GetByID(c.Request.Context(), req.MemberID)
	if err != nil {
		if err.Error() == fmt.Sprintf("member not found: %s", req.MemberID) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "MEMBER_NOT_FOUND",
					"message": "Member not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "MEMBER_RETRIEVAL_FAILED",
				"message": "Failed to retrieve member",
				"details": err.Error(),
			},
		})
		return
	}

	// Verify member belongs to the group
	if member.GroupID != groupID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "MEMBER_NOT_IN_GROUP",
				"message": "Member does not belong to this group",
			},
		})
		return
	}

	// Create bucket list item
	item := &models.BucketListItem{
		ID:          uuid.New().String(),
		GroupID:     groupID,
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CompletedBy: nil,
		CompletedAt: nil,
		CreatedBy:   req.MemberID,
		CreatedAt:   time.Now(),
	}

	// Save item to database
	if err := h.repos.BucketItems().Create(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "ITEM_CREATION_FAILED",
				"message": "Failed to create bucket list item",
				"details": err.Error(),
			},
		})
		return
	}

	// Return created item
	c.JSON(http.StatusCreated, gin.H{
		"item": item,
	})
}

// ToggleCompletion handles PATCH /api/items/:id/complete
func (h *BucketItemHandler) ToggleCompletion(c *gin.Context) {
	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "MISSING_ITEM_ID",
				"message": "Item ID is required",
			},
		})
		return
	}

	// Validate UUID format
	validation := models.ValidateUUID(itemID)
	if !validation.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ITEM_ID",
				"message": "Invalid item ID format",
				"details": validation.Errors,
			},
		})
		return
	}

	var req models.ToggleCompletionRequest
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

	// Validate member ID format
	memberValidation := models.ValidateUUID(req.MemberID)
	if !memberValidation.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_MEMBER_ID",
				"message": "Invalid member ID format",
				"details": memberValidation.Errors,
			},
		})
		return
	}

	// Check if item exists
	item, err := h.repos.BucketItems().GetByID(c.Request.Context(), itemID)
	if err != nil {
		if err.Error() == fmt.Sprintf("bucket item not found: %s", itemID) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "ITEM_NOT_FOUND",
					"message": "Bucket list item not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "ITEM_RETRIEVAL_FAILED",
				"message": "Failed to retrieve bucket list item",
				"details": err.Error(),
			},
		})
		return
	}

	// Check if member exists and belongs to the same group as the item
	member, err := h.repos.Members().GetByID(c.Request.Context(), req.MemberID)
	if err != nil {
		if err.Error() == fmt.Sprintf("member not found: %s", req.MemberID) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "MEMBER_NOT_FOUND",
					"message": "Member not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "MEMBER_RETRIEVAL_FAILED",
				"message": "Failed to retrieve member",
				"details": err.Error(),
			},
		})
		return
	}

	// Verify member belongs to the same group as the item
	if member.GroupID != item.GroupID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "MEMBER_NOT_IN_GROUP",
				"message": "Member does not belong to the same group as this item",
			},
		})
		return
	}

	// Toggle completion status
	if err := h.repos.BucketItems().ToggleCompletion(c.Request.Context(), itemID, req.MemberID, req.Completed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "COMPLETION_TOGGLE_FAILED",
				"message": "Failed to toggle completion status",
				"details": err.Error(),
			},
		})
		return
	}

	// Get updated item to return
	updatedItem, err := h.repos.BucketItems().GetByID(c.Request.Context(), itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "UPDATED_ITEM_RETRIEVAL_FAILED",
				"message": "Failed to retrieve updated item",
				"details": err.Error(),
			},
		})
		return
	}

	// Return updated item
	c.JSON(http.StatusOK, gin.H{
		"item": updatedItem,
	})
}