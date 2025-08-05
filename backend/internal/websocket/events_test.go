package websocket

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"collaborative-bucket-list/internal/models"
	"collaborative-bucket-list/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository interfaces
type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) Create(ctx context.Context, group *models.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepository) GetByID(ctx context.Context, id string) (*models.Group, error) {
	args := m.Called(ctx, id)
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
	return args.Get(0).(*models.Member), args.Error(1)
}

type MockBucketItemRepository struct {
	mock.Mock
}

func (m *MockBucketItemRepository) Create(ctx context.Context, item *models.BucketListItem) error {
	args := m.Called(ctx, item)
	if args.Error(0) == nil {
		// Set ID for successful creation
		item.ID = "test-item-id"
	}
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

// Mock client for testing
type MockClient struct {
	*Client
	sentMessages [][]byte
}

func NewMockClient(roomID, memberID string) *MockClient {
	return &MockClient{
		Client: &Client{
			roomID:   roomID,
			memberID: memberID,
			send:     make(chan []byte, 256),
		},
		sentMessages: make([][]byte, 0),
	}
}

// Mock hub for testing
type MockHub struct {
	broadcastedMessages []BroadcastMessage
}

type BroadcastMessage struct {
	RoomID      string
	MessageType string
	Data        interface{}
}

func (h *MockHub) BroadcastToRoom(roomID string, messageType string, data interface{}) {
	h.broadcastedMessages = append(h.broadcastedMessages, BroadcastMessage{
		RoomID:      roomID,
		MessageType: messageType,
		Data:        data,
	})
}

// Ensure MockHub implements HubInterface
var _ HubInterface = (*MockHub)(nil)

func TestEventHandler_ProcessMessage_InvalidJSON(t *testing.T) {
	mockRepos := NewMockRepositoryManager()
	mockHub := &MockHub{}
	eventHandler := NewEventHandler(mockHub, mockRepos)

	client := NewMockClient("test-room", "test-member")

	// Test invalid JSON
	invalidJSON := []byte(`{"invalid": json}`)
	eventHandler.ProcessMessage(client.Client, invalidJSON)

	// Should not broadcast anything for invalid JSON
	assert.Empty(t, mockHub.broadcastedMessages)
}

func TestEventHandler_HandleJoinGroup_Success(t *testing.T) {
	mockRepos := NewMockRepositoryManager()
	mockHub := &MockHub{}
	eventHandler := NewEventHandler(mockHub, mockRepos)

	client := NewMockClient("test-group-id", "test-member-id")

	// Mock member data
	member := &models.Member{
		ID:        "test-member-id",
		GroupID:   "test-group-id",
		Name:      "Test Member",
		JoinedAt:  time.Now(),
		IsCreator: false,
	}

	mockRepos.members.On("GetByID", mock.Anything, "test-member-id").Return(member, nil)

	// Create join-group message
	payload := JoinGroupPayload{
		GroupID:  "test-group-id",
		MemberID: "test-member-id",
	}

	message := Message{
		Type:     EventJoinGroup,
		RoomID:   "test-group-id",
		MemberID: "test-member-id",
		Data:     payload,
	}

	messageBytes, _ := json.Marshal(message)
	eventHandler.ProcessMessage(client.Client, messageBytes)

	// Verify member-joined event was broadcasted
	assert.Len(t, mockHub.broadcastedMessages, 1)
	assert.Equal(t, EventMemberJoined, mockHub.broadcastedMessages[0].MessageType)
	assert.Equal(t, "test-group-id", mockHub.broadcastedMessages[0].RoomID)
	assert.Equal(t, member, mockHub.broadcastedMessages[0].Data)

	mockRepos.members.AssertExpectations(t)
}

func TestEventHandler_HandleJoinGroup_MemberNotFound(t *testing.T) {
	mockRepos := NewMockRepositoryManager()
	mockHub := &MockHub{}
	eventHandler := NewEventHandler(mockHub, mockRepos)

	client := NewMockClient("test-group-id", "test-member-id")

	mockRepos.members.On("GetByID", mock.Anything, "test-member-id").Return((*models.Member)(nil), errors.New("member not found"))

	payload := JoinGroupPayload{
		GroupID:  "test-group-id",
		MemberID: "test-member-id",
	}

	message := Message{
		Type:     EventJoinGroup,
		RoomID:   "test-group-id",
		MemberID: "test-member-id",
		Data:     payload,
	}

	messageBytes, _ := json.Marshal(message)
	eventHandler.ProcessMessage(client.Client, messageBytes)

	// Should not broadcast anything for member not found
	assert.Empty(t, mockHub.broadcastedMessages)

	mockRepos.members.AssertExpectations(t)
}

func TestEventHandler_HandleAddItem_Success(t *testing.T) {
	mockRepos := NewMockRepositoryManager()
	mockHub := &MockHub{}
	eventHandler := NewEventHandler(mockHub, mockRepos)

	client := NewMockClient("test-group-id", "test-member-id")

	// Mock member data
	member := &models.Member{
		ID:        "test-member-id",
		GroupID:   "test-group-id",
		Name:      "Test Member",
		JoinedAt:  time.Now(),
		IsCreator: false,
	}

	mockRepos.members.On("GetByID", mock.Anything, "test-member-id").Return(member, nil)
	mockRepos.bucketItems.On("Create", mock.Anything, mock.AnythingOfType("*models.BucketListItem")).Return(nil)

	// Create add-item message
	itemRequest := models.CreateItemRequest{
		Title:       "Test Item",
		Description: stringPtr("Test Description"),
		MemberID:    "test-member-id",
	}

	payload := AddItemPayload{
		GroupID: "test-group-id",
		Item:    itemRequest,
	}

	message := Message{
		Type:     EventAddItem,
		RoomID:   "test-group-id",
		MemberID: "test-member-id",
		Data:     payload,
	}

	messageBytes, _ := json.Marshal(message)
	eventHandler.ProcessMessage(client.Client, messageBytes)

	// Verify item-added event was broadcasted
	assert.Len(t, mockHub.broadcastedMessages, 1)
	assert.Equal(t, EventItemAdded, mockHub.broadcastedMessages[0].MessageType)
	assert.Equal(t, "test-group-id", mockHub.broadcastedMessages[0].RoomID)

	// Verify the broadcasted item has correct data
	broadcastedItem := mockHub.broadcastedMessages[0].Data.(*models.BucketListItem)
	assert.Equal(t, "Test Item", broadcastedItem.Title)
	assert.Equal(t, "Test Description", *broadcastedItem.Description)
	assert.Equal(t, "test-group-id", broadcastedItem.GroupID)
	assert.Equal(t, "test-member-id", broadcastedItem.CreatedBy)
	assert.False(t, broadcastedItem.Completed)

	mockRepos.members.AssertExpectations(t)
	mockRepos.bucketItems.AssertExpectations(t)
}

func TestEventHandler_HandleToggleCompletion_Success(t *testing.T) {
	mockRepos := NewMockRepositoryManager()
	mockHub := &MockHub{}
	eventHandler := NewEventHandler(mockHub, mockRepos)

	client := NewMockClient("test-group-id", "test-member-id")

	// Mock member data
	member := &models.Member{
		ID:        "test-member-id",
		GroupID:   "test-group-id",
		Name:      "Test Member",
		JoinedAt:  time.Now(),
		IsCreator: false,
	}

	// Mock item data
	item := &models.BucketListItem{
		ID:          "test-item-id",
		GroupID:     "test-group-id",
		Title:       "Test Item",
		Description: stringPtr("Test Description"),
		Completed:   false,
		CreatedBy:   "test-member-id",
		CreatedAt:   time.Now(),
	}

	updatedItem := &models.BucketListItem{
		ID:          "test-item-id",
		GroupID:     "test-group-id",
		Title:       "Test Item",
		Description: stringPtr("Test Description"),
		Completed:   true,
		CompletedBy: stringPtr("test-member-id"),
		CompletedAt: timePtr(time.Now()),
		CreatedBy:   "test-member-id",
		CreatedAt:   time.Now(),
	}

	mockRepos.members.On("GetByID", mock.Anything, "test-member-id").Return(member, nil)
	mockRepos.bucketItems.On("GetByID", mock.Anything, "test-item-id").Return(item, nil).Once()
	mockRepos.bucketItems.On("ToggleCompletion", mock.Anything, "test-item-id", "test-member-id", true).Return(nil)
	mockRepos.bucketItems.On("GetByID", mock.Anything, "test-item-id").Return(updatedItem, nil).Once()

	// Create toggle-completion message
	payload := ToggleCompletionPayload{
		GroupID:   "test-group-id",
		ItemID:    "test-item-id",
		Completed: true,
		MemberID:  "test-member-id",
	}

	message := Message{
		Type:     EventToggleCompletion,
		RoomID:   "test-group-id",
		MemberID: "test-member-id",
		Data:     payload,
	}

	messageBytes, _ := json.Marshal(message)
	eventHandler.ProcessMessage(client.Client, messageBytes)

	// Verify item-updated event was broadcasted
	assert.Len(t, mockHub.broadcastedMessages, 1)
	assert.Equal(t, EventItemUpdated, mockHub.broadcastedMessages[0].MessageType)
	assert.Equal(t, "test-group-id", mockHub.broadcastedMessages[0].RoomID)

	// Verify the broadcasted item has updated completion status
	broadcastedItem := mockHub.broadcastedMessages[0].Data.(*models.BucketListItem)
	assert.True(t, broadcastedItem.Completed)
	assert.Equal(t, "test-member-id", *broadcastedItem.CompletedBy)
	assert.NotNil(t, broadcastedItem.CompletedAt)

	mockRepos.members.AssertExpectations(t)
	mockRepos.bucketItems.AssertExpectations(t)
}

func TestEventHandler_HandleToggleCompletion_ItemNotFound(t *testing.T) {
	mockRepos := NewMockRepositoryManager()
	mockHub := &MockHub{}
	eventHandler := NewEventHandler(mockHub, mockRepos)

	client := NewMockClient("test-group-id", "test-member-id")

	// Mock member data
	member := &models.Member{
		ID:        "test-member-id",
		GroupID:   "test-group-id",
		Name:      "Test Member",
		JoinedAt:  time.Now(),
		IsCreator: false,
	}

	mockRepos.members.On("GetByID", mock.Anything, "test-member-id").Return(member, nil)
	mockRepos.bucketItems.On("GetByID", mock.Anything, "test-item-id").Return((*models.BucketListItem)(nil), errors.New("item not found"))

	payload := ToggleCompletionPayload{
		GroupID:   "test-group-id",
		ItemID:    "test-item-id",
		Completed: true,
		MemberID:  "test-member-id",
	}

	message := Message{
		Type:     EventToggleCompletion,
		RoomID:   "test-group-id",
		MemberID: "test-member-id",
		Data:     payload,
	}

	messageBytes, _ := json.Marshal(message)
	eventHandler.ProcessMessage(client.Client, messageBytes)

	// Should not broadcast anything for item not found
	assert.Empty(t, mockHub.broadcastedMessages)

	mockRepos.members.AssertExpectations(t)
	mockRepos.bucketItems.AssertExpectations(t)
}

func TestEventHandler_ProcessMessage_RoomMismatch(t *testing.T) {
	mockRepos := NewMockRepositoryManager()
	mockHub := &MockHub{}
	eventHandler := NewEventHandler(mockHub, mockRepos)

	client := NewMockClient("test-group-id", "test-member-id")

	// Create message with different room ID
	message := Message{
		Type:     EventJoinGroup,
		RoomID:   "different-group-id",
		MemberID: "test-member-id",
		Data:     JoinGroupPayload{},
	}

	messageBytes, _ := json.Marshal(message)
	eventHandler.ProcessMessage(client.Client, messageBytes)

	// Should not broadcast anything for room mismatch
	assert.Empty(t, mockHub.broadcastedMessages)
}

func TestEventHandler_ProcessMessage_MemberMismatch(t *testing.T) {
	mockRepos := NewMockRepositoryManager()
	mockHub := &MockHub{}
	eventHandler := NewEventHandler(mockHub, mockRepos)

	client := NewMockClient("test-group-id", "test-member-id")

	// Create message with different member ID
	message := Message{
		Type:     EventJoinGroup,
		RoomID:   "test-group-id",
		MemberID: "different-member-id",
		Data:     JoinGroupPayload{},
	}

	messageBytes, _ := json.Marshal(message)
	eventHandler.ProcessMessage(client.Client, messageBytes)

	// Should not broadcast anything for member mismatch
	assert.Empty(t, mockHub.broadcastedMessages)
}

func TestEventHandler_ProcessMessage_UnknownEventType(t *testing.T) {
	mockRepos := NewMockRepositoryManager()
	mockHub := &MockHub{}
	eventHandler := NewEventHandler(mockHub, mockRepos)

	client := NewMockClient("test-group-id", "test-member-id")

	// Create message with unknown event type
	message := Message{
		Type:     "unknown-event",
		RoomID:   "test-group-id",
		MemberID: "test-member-id",
		Data:     map[string]interface{}{},
	}

	messageBytes, _ := json.Marshal(message)
	eventHandler.ProcessMessage(client.Client, messageBytes)

	// Should not broadcast anything for unknown event type
	assert.Empty(t, mockHub.broadcastedMessages)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}