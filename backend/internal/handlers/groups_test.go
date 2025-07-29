package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"collaborative-bucket-list/internal/middleware"
	"collaborative-bucket-list/internal/models"
	"collaborative-bucket-list/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository implementations for testing
type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) Create(ctx context.Context, group *models.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepository) GetByID(ctx context.Context, id string) (*models.Group, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupRepository) GetByUserID(ctx context.Context, userID string) ([]models.Group, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Group), args.Error(1)
}

func (m *MockGroupRepository) Update(ctx context.Context, group *models.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGroupRepository) GetWithDetails(ctx context.Context, id string) (*models.GroupWithDetails, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GroupWithDetails), args.Error(1)
}

func (m *MockGroupRepository) GetSummariesByUserID(ctx context.Context, userID string) ([]models.GroupSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.GroupSummary), args.Error(1)
}

type MockMemberRepository struct {
	mock.Mock
}

func (m *MockMemberRepository) Create(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockMemberRepository) GetByID(ctx context.Context, id string) (*models.Member, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Member), args.Error(1)
}

func (m *MockMemberRepository) GetByGroupID(ctx context.Context, groupID string) ([]models.Member, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]models.Member), args.Error(1)
}

func (m *MockMemberRepository) GetByUserID(ctx context.Context, userID string) ([]models.Member, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.Member), args.Error(1)
}

func (m *MockMemberRepository) Update(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockMemberRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMemberRepository) ExistsByGroupAndUser(ctx context.Context, groupID, userID string) (bool, error) {
	args := m.Called(ctx, groupID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockMemberRepository) GetCreatorByGroupID(ctx context.Context, groupID string) (*models.Member, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Member), args.Error(1)
}

type MockBucketItemRepository struct {
	mock.Mock
}

func (m *MockBucketItemRepository) Create(ctx context.Context, item *models.BucketListItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockBucketItemRepository) GetByID(ctx context.Context, id string) (*models.BucketListItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BucketListItem), args.Error(1)
}

func (m *MockBucketItemRepository) GetByGroupID(ctx context.Context, groupID string) ([]models.BucketListItem, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]models.BucketListItem), args.Error(1)
}

func (m *MockBucketItemRepository) Update(ctx context.Context, item *models.BucketListItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockBucketItemRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBucketItemRepository) ToggleCompletion(ctx context.Context, itemID, memberID string, completed bool) error {
	args := m.Called(ctx, itemID, memberID, completed)
	return args.Error(0)
}

func (m *MockBucketItemRepository) GetCompletionStats(ctx context.Context, groupID string) (total, completed int, err error) {
	args := m.Called(ctx, groupID)
	return args.Int(0), args.Int(1), args.Error(2)
}

type MockRepositoryManager struct {
	mock.Mock
	groups      *MockGroupRepository
	members     *MockMemberRepository
	bucketItems *MockBucketItemRepository
}

func NewMockRepositoryManager() *MockRepositoryManager {
	return &MockRepositoryManager{
		groups:      &MockGroupRepository{},
		members:     &MockMemberRepository{},
		bucketItems: &MockBucketItemRepository{},
	}
}

func (m *MockRepositoryManager) Groups() repositories.GroupRepository {
	return m.groups
}

func (m *MockRepositoryManager) Members() repositories.MemberRepository {
	return m.members
}

func (m *MockRepositoryManager) BucketItems() repositories.BucketItemRepository {
	return m.bucketItems
}

func (m *MockRepositoryManager) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// Test helper functions
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createTestUser() *middleware.SupabaseUser {
	return &middleware.SupabaseUser{
		ID:    uuid.New().String(),
		Email: "test@example.com",
		Role:  "authenticated",
	}
}

func addUserToContext(c *gin.Context, user *middleware.SupabaseUser) {
	c.Set("user", user)
	c.Set("userID", user.ID)
	c.Set("userEmail", user.Email)
}

func TestGroupHandler_CreateGroup(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockRepositoryManager)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful group creation",
			requestBody: models.CreateGroupRequest{
				Name:     "Test Group",
				Deadline: nil,
			},
			setupMocks: func(m *MockRepositoryManager) {
				m.On("WithTx", mock.Anything, mock.AnythingOfType("func(*sql.Tx) error")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid request body",
			requestBody: models.CreateGroupRequest{
				Name: "a", // Too short name should fail validation
			},
			setupMocks:     func(m *MockRepositoryManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "malformed JSON",
			requestBody:    "invalid json",
			setupMocks:     func(m *MockRepositoryManager) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST_BODY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepos := NewMockRepositoryManager()
			tt.setupMocks(mockRepos)
			
			handler := NewGroupHandler(mockRepos)
			router := setupTestRouter()
			
			// Add middleware to set user context
			router.Use(func(c *gin.Context) {
				user := createTestUser()
				addUserToContext(c, user)
				c.Next()
			})
			
			router.POST("/groups", handler.CreateGroup)

			// Create request
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/groups", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				errorObj, exists := response["error"].(map[string]interface{})
				assert.True(t, exists)
				assert.Equal(t, tt.expectedError, errorObj["code"])
			}

			mockRepos.AssertExpectations(t)
		})
	}
}

func TestGroupHandler_GetGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupID        string
		setupMocks     func(*MockRepositoryManager, string)
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "successful group retrieval",
			groupID: uuid.New().String(),
			setupMocks: func(m *MockRepositoryManager, groupID string) {
				groupDetails := &models.GroupWithDetails{
					Group: models.Group{
						ID:        groupID,
						Name:      "Test Group",
						CreatedAt: time.Now(),
						CreatedBy: uuid.New().String(),
					},
					Members: []models.Member{},
					Items:   []models.BucketListItem{},
				}
				m.groups.On("GetWithDetails", mock.Anything, groupID).Return(groupDetails, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "group not found",
			groupID: uuid.New().String(),
			setupMocks: func(m *MockRepositoryManager, groupID string) {
				m.groups.On("GetWithDetails", mock.Anything, groupID).Return(nil, fmt.Errorf("group not found: %s", groupID))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "GROUP_NOT_FOUND",
		},
		{
			name:           "invalid group ID format",
			groupID:        "invalid-uuid",
			setupMocks:     func(m *MockRepositoryManager, groupID string) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_GROUP_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepos := NewMockRepositoryManager()
			tt.setupMocks(mockRepos, tt.groupID)
			
			handler := NewGroupHandler(mockRepos)
			router := setupTestRouter()
			router.GET("/groups/:id", handler.GetGroup)

			// Create request
			req, err := http.NewRequest("GET", "/groups/"+tt.groupID, nil)
			assert.NoError(t, err)

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				errorObj, exists := response["error"].(map[string]interface{})
				assert.True(t, exists)
				assert.Equal(t, tt.expectedError, errorObj["code"])
			}

			mockRepos.AssertExpectations(t)
		})
	}
}

func TestGroupHandler_GetUserGroups(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockRepositoryManager, string)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful user groups retrieval",
			setupMocks: func(m *MockRepositoryManager, userID string) {
				summaries := []models.GroupSummary{
					{
						Group: models.Group{
							ID:        uuid.New().String(),
							Name:      "Test Group 1",
							CreatedAt: time.Now(),
							CreatedBy: userID,
						},
						MemberCount:     2,
						ItemCount:       5,
						CompletedCount:  2,
						ProgressPercent: 40.0,
					},
				}
				m.groups.On("GetSummariesByUserID", mock.Anything, userID).Return(summaries, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "database error",
			setupMocks: func(m *MockRepositoryManager, userID string) {
				m.groups.On("GetSummariesByUserID", mock.Anything, userID).Return([]models.GroupSummary{}, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "USER_GROUPS_RETRIEVAL_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepos := NewMockRepositoryManager()
			user := createTestUser()
			tt.setupMocks(mockRepos, user.ID)
			
			handler := NewGroupHandler(mockRepos)
			router := setupTestRouter()
			
			// Add middleware to set user context
			router.Use(func(c *gin.Context) {
				addUserToContext(c, user)
				c.Next()
			})
			
			router.GET("/users/groups", handler.GetUserGroups)

			// Create request
			req, err := http.NewRequest("GET", "/users/groups", nil)
			assert.NoError(t, err)

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				errorObj, exists := response["error"].(map[string]interface{})
				assert.True(t, exists)
				assert.Equal(t, tt.expectedError, errorObj["code"])
			}

			mockRepos.AssertExpectations(t)
		})
	}
}