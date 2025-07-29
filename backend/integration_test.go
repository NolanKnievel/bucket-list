package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"collaborative-bucket-list/internal/handlers"
	"collaborative-bucket-list/internal/middleware"
	"collaborative-bucket-list/internal/models"
	"collaborative-bucket-list/internal/repositories"
	"collaborative-bucket-list/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*gin.Engine, repositories.RepositoryManager) {
	// Set up test database connection
	dbConfig := &database.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "password123",
		DBName:   "collaborative_bucket_list",
		SSLMode:  "disable",
	}

	err := database.Connect(dbConfig)
	require.NoError(t, err)

	// Run migrations
	err = database.RunMigrations("migrations")
	require.NoError(t, err)

	// Initialize repository manager
	repoManager := repositories.NewPostgresRepositoryManager(database.DB)

	// Initialize handlers
	groupHandler := handlers.NewGroupHandler(repoManager)

	// Set up Gin router
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add test middleware to set user context
	r.Use(func(c *gin.Context) {
		// For testing, we'll set a mock user
		user := &middleware.SupabaseUser{
			ID:    uuid.New().String(),
			Email: "test@example.com",
			Role:  "authenticated",
		}
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("userEmail", user.Email)
		c.Next()
	})

	// Set up routes
	api := r.Group("/api")
	{
		api.POST("/groups", groupHandler.CreateGroup)
		api.GET("/groups/:id", groupHandler.GetGroup)
		api.POST("/groups/:id/join", groupHandler.JoinGroup)
		api.GET("/users/groups", groupHandler.GetUserGroups)
	}

	return r, repoManager
}

func TestJoinGroupIntegration(t *testing.T) {
	router, _ := setupTestServer(t)
	defer database.Close()

	// Step 1: Create a group
	createGroupReq := models.CreateGroupRequest{
		Name: "Integration Test Group",
	}
	body, err := json.Marshal(createGroupReq)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/groups", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	groupID := createResponse["id"].(string)
	assert.NotEmpty(t, groupID)

	// Step 2: Join the group as an anonymous user
	joinGroupReq := models.JoinGroupRequest{
		MemberName: "Anonymous Test User",
		UserID:     nil,
	}
	body, err = json.Marshal(joinGroupReq)
	require.NoError(t, err)

	req, err = http.NewRequest("POST", fmt.Sprintf("/api/groups/%s/join", groupID), bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var joinResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &joinResponse)
	require.NoError(t, err)

	member := joinResponse["member"].(map[string]interface{})
	assert.Equal(t, groupID, member["groupId"])
	assert.Equal(t, "Anonymous Test User", member["name"])
	assert.False(t, member["isCreator"].(bool))
	assert.Nil(t, member["userId"])

	// Step 3: Join the group as an authenticated user
	userID := uuid.New().String()
	joinGroupReq = models.JoinGroupRequest{
		MemberName: "Authenticated Test User",
		UserID:     &userID,
	}
	body, err = json.Marshal(joinGroupReq)
	require.NoError(t, err)

	req, err = http.NewRequest("POST", fmt.Sprintf("/api/groups/%s/join", groupID), bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &joinResponse)
	require.NoError(t, err)

	member = joinResponse["member"].(map[string]interface{})
	assert.Equal(t, groupID, member["groupId"])
	assert.Equal(t, "Authenticated Test User", member["name"])
	assert.False(t, member["isCreator"].(bool))
	assert.Equal(t, userID, member["userId"])

	// Step 4: Try to join the same group again with the same user (should fail)
	req, err = http.NewRequest("POST", fmt.Sprintf("/api/groups/%s/join", groupID), bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	errorObj := errorResponse["error"].(map[string]interface{})
	assert.Equal(t, "ALREADY_MEMBER", errorObj["code"])

	// Step 5: Verify the group now has all members
	req, err = http.NewRequest("GET", fmt.Sprintf("/api/groups/%s", groupID), nil)
	require.NoError(t, err)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var groupResponse models.GroupWithDetails
	err = json.Unmarshal(w.Body.Bytes(), &groupResponse)
	require.NoError(t, err)

	assert.Equal(t, groupID, groupResponse.ID)
	assert.Equal(t, "Integration Test Group", groupResponse.Name)
	assert.Len(t, groupResponse.Members, 3) // Creator + 2 joined members

	// Verify member details
	memberNames := make(map[string]bool)
	creatorCount := 0
	for _, member := range groupResponse.Members {
		memberNames[member.Name] = true
		if member.IsCreator {
			creatorCount++
		}
	}

	assert.True(t, memberNames["test@example.com"]) // Creator (uses email as name)
	assert.True(t, memberNames["Anonymous Test User"])
	assert.True(t, memberNames["Authenticated Test User"])
	assert.Equal(t, 1, creatorCount) // Only one creator

	// Step 6: Test joining non-existent group
	nonExistentGroupID := uuid.New().String()
	joinGroupReq = models.JoinGroupRequest{
		MemberName: "Test User",
		UserID:     nil,
	}
	body, err = json.Marshal(joinGroupReq)
	require.NoError(t, err)

	req, err = http.NewRequest("POST", fmt.Sprintf("/api/groups/%s/join", nonExistentGroupID), bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	errorObj = errorResponse["error"].(map[string]interface{})
	assert.Equal(t, "GROUP_NOT_FOUND", errorObj["code"])
}