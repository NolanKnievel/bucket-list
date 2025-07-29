package handlers

import (
	"context"
	"database/sql"

	"collaborative-bucket-list/internal/models"
	"collaborative-bucket-list/internal/repositories"

	"github.com/stretchr/testify/mock"
)

// Mock repositories for testing
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