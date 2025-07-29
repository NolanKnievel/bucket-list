package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"collaborative-bucket-list/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock definitions are in mocks_test.go

func TestBucketItemHandler_CreateItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful item creation", func(t *testing.T) {
		// Setup
		mockRepoManager := NewMockRepositoryManager()
		groupID := uuid.New().String()
		memberID := uuid.New().String()
		
		group := &models.Group{
			ID:        groupID,
			Name:      "Test Group",
			CreatedAt: time.Now(),
			CreatedBy: uuid.New().String(),
		}
		member := &models.Member{
			ID:      memberID,
			GroupID: groupID,
			Name:    "Test Member",
		}
		
		mockRepoManager.groups.On("GetByID", mock.Anything, groupID).Return(group, nil)
		mockRepoManager.members.On("GetByID", mock.Anything, memberID).Return(member, nil)
		mockRepoManager.bucketItems.On("Create", mock.Anything, mock.AnythingOfType("*models.BucketListItem")).Return(nil)
		
		handler := NewBucketItemHandler(mockRepoManager)
		
		requestBody := models.CreateItemRequest{
			Title:       "Visit Paris",
			Description: stringPtr("See the Eiffel Tower"),
			MemberID:    memberID,
		}
		
		// Create request
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/groups/%s/items", groupID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		// Create response recorder
		w := httptest.NewRecorder()
		
		// Create Gin context
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: groupID}}
		
		// Execute
		handler.CreateItem(c)
		
		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
		
		// Verify mocks
		mockRepoManager.groups.AssertExpectations(t)
		mockRepoManager.members.AssertExpectations(t)
		mockRepoManager.bucketItems.AssertExpectations(t)
	})

	t.Run("invalid group ID format", func(t *testing.T) {
		mockRepoManager := NewMockRepositoryManager()
		handler := NewBucketItemHandler(mockRepoManager)
		
		requestBody := models.CreateItemRequest{
			Title:    "Visit Paris",
			MemberID: uuid.New().String(),
		}
		
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/groups/invalid-uuid/items", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
		
		handler.CreateItem(c)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		errorObj, exists := response["error"].(map[string]interface{})
		assert.True(t, exists)
		assert.Equal(t, "INVALID_GROUP_ID", errorObj["code"])
	})

	t.Run("missing title", func(t *testing.T) {
		mockRepoManager := NewMockRepositoryManager()
		handler := NewBucketItemHandler(mockRepoManager)
		
		requestBody := map[string]interface{}{
			"memberId": uuid.New().String(),
		}
		
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/groups/%s/items", uuid.New().String()), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
		
		handler.CreateItem(c)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		errorObj, exists := response["error"].(map[string]interface{})
		assert.True(t, exists)
		assert.Equal(t, "INVALID_REQUEST_BODY", errorObj["code"])
	})

	t.Run("group not found", func(t *testing.T) {
		mockRepoManager := NewMockRepositoryManager()
		groupID := uuid.New().String()
		
		mockRepoManager.groups.On("GetByID", mock.Anything, groupID).Return(nil, fmt.Errorf("group not found: %s", groupID))
		
		handler := NewBucketItemHandler(mockRepoManager)
		
		requestBody := models.CreateItemRequest{
			Title:    "Visit Paris",
			MemberID: uuid.New().String(),
		}
		
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/groups/%s/items", groupID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: groupID}}
		
		handler.CreateItem(c)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		errorObj, exists := response["error"].(map[string]interface{})
		assert.True(t, exists)
		assert.Equal(t, "GROUP_NOT_FOUND", errorObj["code"])
		
		mockRepoManager.groups.AssertExpectations(t)
	})

	t.Run("member not found", func(t *testing.T) {
		mockRepoManager := NewMockRepositoryManager()
		groupID := uuid.New().String()
		memberID := uuid.New().String()
		
		group := &models.Group{
			ID:        groupID,
			Name:      "Test Group",
			CreatedAt: time.Now(),
			CreatedBy: uuid.New().String(),
		}
		
		mockRepoManager.groups.On("GetByID", mock.Anything, groupID).Return(group, nil)
		mockRepoManager.members.On("GetByID", mock.Anything, memberID).Return(nil, fmt.Errorf("member not found: %s", memberID))
		
		handler := NewBucketItemHandler(mockRepoManager)
		
		requestBody := models.CreateItemRequest{
			Title:    "Visit Paris",
			MemberID: memberID,
		}
		
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/groups/%s/items", groupID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: groupID}}
		
		handler.CreateItem(c)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		errorObj, exists := response["error"].(map[string]interface{})
		assert.True(t, exists)
		assert.Equal(t, "MEMBER_NOT_FOUND", errorObj["code"])
		
		mockRepoManager.groups.AssertExpectations(t)
		mockRepoManager.members.AssertExpectations(t)
	})

	t.Run("member not in group", func(t *testing.T) {
		mockRepoManager := NewMockRepositoryManager()
		groupID := uuid.New().String()
		memberID := uuid.New().String()
		
		group := &models.Group{
			ID:        groupID,
			Name:      "Test Group",
			CreatedAt: time.Now(),
			CreatedBy: uuid.New().String(),
		}
		member := &models.Member{
			ID:      memberID,
			GroupID: uuid.New().String(), // Different group ID
			Name:    "Test Member",
		}
		
		mockRepoManager.groups.On("GetByID", mock.Anything, groupID).Return(group, nil)
		mockRepoManager.members.On("GetByID", mock.Anything, memberID).Return(member, nil)
		
		handler := NewBucketItemHandler(mockRepoManager)
		
		requestBody := models.CreateItemRequest{
			Title:    "Visit Paris",
			MemberID: memberID,
		}
		
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/groups/%s/items", groupID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: groupID}}
		
		handler.CreateItem(c)
		
		assert.Equal(t, http.StatusForbidden, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		errorObj, exists := response["error"].(map[string]interface{})
		assert.True(t, exists)
		assert.Equal(t, "MEMBER_NOT_IN_GROUP", errorObj["code"])
		
		mockRepoManager.groups.AssertExpectations(t)
		mockRepoManager.members.AssertExpectations(t)
	})
}

func TestBucketItemHandler_ToggleCompletion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful completion toggle", func(t *testing.T) {
		mockRepoManager := NewMockRepositoryManager()
		itemID := uuid.New().String()
		memberID := uuid.New().String()
		groupID := uuid.New().String()
		
		item := &models.BucketListItem{
			ID:        itemID,
			GroupID:   groupID,
			Title:     "Visit Paris",
			Completed: false,
			CreatedAt: time.Now(),
		}
		member := &models.Member{
			ID:      memberID,
			GroupID: groupID,
			Name:    "Test Member",
		}
		updatedItem := &models.BucketListItem{
			ID:          itemID,
			GroupID:     groupID,
			Title:       "Visit Paris",
			Completed:   true,
			CompletedBy: &memberID,
			CompletedAt: timePtr(time.Now()),
			CreatedAt:   item.CreatedAt,
		}
		
		mockRepoManager.bucketItems.On("GetByID", mock.Anything, itemID).Return(item, nil).Once()
		mockRepoManager.members.On("GetByID", mock.Anything, memberID).Return(member, nil)
		mockRepoManager.bucketItems.On("ToggleCompletion", mock.Anything, itemID, memberID, true).Return(nil)
		mockRepoManager.bucketItems.On("GetByID", mock.Anything, itemID).Return(updatedItem, nil).Once()
		
		handler := NewBucketItemHandler(mockRepoManager)
		
		requestBody := models.ToggleCompletionRequest{
			Completed: true,
			MemberID:  memberID,
		}
		
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/items/%s/complete", itemID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: itemID}}
		
		handler.ToggleCompletion(c)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		mockRepoManager.bucketItems.AssertExpectations(t)
		mockRepoManager.members.AssertExpectations(t)
	})

	t.Run("invalid item ID format", func(t *testing.T) {
		mockRepoManager := NewMockRepositoryManager()
		handler := NewBucketItemHandler(mockRepoManager)
		
		requestBody := models.ToggleCompletionRequest{
			Completed: true,
			MemberID:  uuid.New().String(),
		}
		
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPatch, "/api/items/invalid-uuid/complete", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
		
		handler.ToggleCompletion(c)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		errorObj, exists := response["error"].(map[string]interface{})
		assert.True(t, exists)
		assert.Equal(t, "INVALID_ITEM_ID", errorObj["code"])
	})

	t.Run("item not found", func(t *testing.T) {
		mockRepoManager := NewMockRepositoryManager()
		itemID := uuid.New().String()
		
		mockRepoManager.bucketItems.On("GetByID", mock.Anything, itemID).Return(nil, fmt.Errorf("bucket item not found: %s", itemID))
		
		handler := NewBucketItemHandler(mockRepoManager)
		
		requestBody := models.ToggleCompletionRequest{
			Completed: true,
			MemberID:  uuid.New().String(),
		}
		
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/items/%s/complete", itemID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: itemID}}
		
		handler.ToggleCompletion(c)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		errorObj, exists := response["error"].(map[string]interface{})
		assert.True(t, exists)
		assert.Equal(t, "ITEM_NOT_FOUND", errorObj["code"])
		
		mockRepoManager.bucketItems.AssertExpectations(t)
	})

	t.Run("member not in same group as item", func(t *testing.T) {
		mockRepoManager := NewMockRepositoryManager()
		itemID := uuid.New().String()
		memberID := uuid.New().String()
		
		item := &models.BucketListItem{
			ID:        itemID,
			GroupID:   uuid.New().String(),
			Title:     "Visit Paris",
			Completed: false,
			CreatedAt: time.Now(),
		}
		member := &models.Member{
			ID:      memberID,
			GroupID: uuid.New().String(), // Different group ID
			Name:    "Test Member",
		}
		
		mockRepoManager.bucketItems.On("GetByID", mock.Anything, itemID).Return(item, nil)
		mockRepoManager.members.On("GetByID", mock.Anything, memberID).Return(member, nil)
		
		handler := NewBucketItemHandler(mockRepoManager)
		
		requestBody := models.ToggleCompletionRequest{
			Completed: true,
			MemberID:  memberID,
		}
		
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/items/%s/complete", itemID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: itemID}}
		
		handler.ToggleCompletion(c)
		
		assert.Equal(t, http.StatusForbidden, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		errorObj, exists := response["error"].(map[string]interface{})
		assert.True(t, exists)
		assert.Equal(t, "MEMBER_NOT_IN_GROUP", errorObj["code"])
		
		mockRepoManager.bucketItems.AssertExpectations(t)
		mockRepoManager.members.AssertExpectations(t)
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}