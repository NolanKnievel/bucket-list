package repositories

import (
	"context"
	"testing"
	"time"

	"collaborative-bucket-list/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestMember creates a test member for use in tests
func createTestMember(groupID string) *models.Member {
	userID := uuid.New().String()
	return &models.Member{
		ID:        uuid.New().String(),
		GroupID:   groupID,
		UserID:    &userID,
		Name:      "Test Member",
		JoinedAt:  time.Now(),
		IsCreator: false,
	}
}

func TestPostgresMemberRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	ctx := context.Background()
	
	// Create a test group first
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	t.Run("successful creation", func(t *testing.T) {
		member := createTestMember(group.ID)
		
		err := memberRepo.Create(ctx, member)
		assert.NoError(t, err)
		
		// Verify the member was created
		retrieved, err := memberRepo.GetByID(ctx, member.ID)
		require.NoError(t, err)
		assert.Equal(t, member.ID, retrieved.ID)
		assert.Equal(t, member.Name, retrieved.Name)
		assert.Equal(t, member.GroupID, retrieved.GroupID)
		assert.Equal(t, *member.UserID, *retrieved.UserID)
	})
	
	t.Run("invalid member data", func(t *testing.T) {
		member := &models.Member{
			ID:       uuid.New().String(),
			GroupID:  group.ID,
			Name:     "", // Invalid: empty name
			JoinedAt: time.Now(),
		}
		
		err := memberRepo.Create(ctx, member)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid member data")
	})
	
	t.Run("member without user ID (anonymous)", func(t *testing.T) {
		member := &models.Member{
			ID:        uuid.New().String(),
			GroupID:   group.ID,
			UserID:    nil, // Anonymous member
			Name:      "Anonymous Member",
			JoinedAt:  time.Now(),
			IsCreator: false,
		}
		
		err := memberRepo.Create(ctx, member)
		assert.NoError(t, err)
		
		retrieved, err := memberRepo.GetByID(ctx, member.ID)
		require.NoError(t, err)
		assert.Nil(t, retrieved.UserID)
		assert.Equal(t, "Anonymous Member", retrieved.Name)
	})
}

func TestPostgresMemberRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	ctx := context.Background()
	
	// Create a test group first
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	t.Run("existing member", func(t *testing.T) {
		member := createTestMember(group.ID)
		err := memberRepo.Create(ctx, member)
		require.NoError(t, err)
		
		retrieved, err := memberRepo.GetByID(ctx, member.ID)
		require.NoError(t, err)
		assert.Equal(t, member.ID, retrieved.ID)
		assert.Equal(t, member.Name, retrieved.Name)
		assert.Equal(t, member.GroupID, retrieved.GroupID)
	})
	
	t.Run("non-existent member", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		
		_, err := memberRepo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member not found")
	})
}

func TestPostgresMemberRepository_GetByGroupID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	ctx := context.Background()
	
	// Create test groups
	group1 := createTestGroup()
	group2 := createTestGroup()
	err := groupRepo.Create(ctx, group1)
	require.NoError(t, err)
	err = groupRepo.Create(ctx, group2)
	require.NoError(t, err)
	
	t.Run("group with members", func(t *testing.T) {
		// Create members for group1
		member1 := createTestMember(group1.ID)
		member1.IsCreator = true
		member2 := createTestMember(group1.ID)
		member2.IsCreator = false
		
		err := memberRepo.Create(ctx, member1)
		require.NoError(t, err)
		err = memberRepo.Create(ctx, member2)
		require.NoError(t, err)
		
		// Create a member for group2
		member3 := createTestMember(group2.ID)
		err = memberRepo.Create(ctx, member3)
		require.NoError(t, err)
		
		members, err := memberRepo.GetByGroupID(ctx, group1.ID)
		require.NoError(t, err)
		assert.Len(t, members, 2)
		
		// Verify all members belong to group1
		for _, member := range members {
			assert.Equal(t, group1.ID, member.GroupID)
		}
		
		// Verify ordering (joined_at ASC)
		assert.True(t, members[0].JoinedAt.Before(members[1].JoinedAt) || members[0].JoinedAt.Equal(members[1].JoinedAt))
	})
	
	t.Run("group with no members", func(t *testing.T) {
		emptyGroup := createTestGroup()
		err := groupRepo.Create(ctx, emptyGroup)
		require.NoError(t, err)
		
		members, err := memberRepo.GetByGroupID(ctx, emptyGroup.ID)
		require.NoError(t, err)
		assert.Empty(t, members)
	})
}

func TestPostgresMemberRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	ctx := context.Background()
	
	userID := uuid.New().String()
	
	// Create test groups
	group1 := createTestGroup()
	group2 := createTestGroup()
	err := groupRepo.Create(ctx, group1)
	require.NoError(t, err)
	err = groupRepo.Create(ctx, group2)
	require.NoError(t, err)
	
	t.Run("user with memberships", func(t *testing.T) {
		// Create memberships for the user
		member1 := createTestMember(group1.ID)
		member1.UserID = &userID
		member2 := createTestMember(group2.ID)
		member2.UserID = &userID
		
		err := memberRepo.Create(ctx, member1)
		require.NoError(t, err)
		err = memberRepo.Create(ctx, member2)
		require.NoError(t, err)
		
		// Create a membership for another user
		otherMember := createTestMember(group1.ID)
		otherUserID := uuid.New().String()
		otherMember.UserID = &otherUserID
		err = memberRepo.Create(ctx, otherMember)
		require.NoError(t, err)
		
		members, err := memberRepo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, members, 2)
		
		// Verify all memberships belong to the user
		for _, member := range members {
			assert.Equal(t, userID, *member.UserID)
		}
	})
	
	t.Run("user with no memberships", func(t *testing.T) {
		nonExistentUserID := uuid.New().String()
		
		members, err := memberRepo.GetByUserID(ctx, nonExistentUserID)
		require.NoError(t, err)
		assert.Empty(t, members)
	})
}

func TestPostgresMemberRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	ctx := context.Background()
	
	// Create a test group first
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	t.Run("successful update", func(t *testing.T) {
		member := createTestMember(group.ID)
		err := memberRepo.Create(ctx, member)
		require.NoError(t, err)
		
		// Update the member
		member.Name = "Updated Member Name"
		newUserID := uuid.New().String()
		member.UserID = &newUserID
		
		err = memberRepo.Update(ctx, member)
		assert.NoError(t, err)
		
		// Verify the update
		updated, err := memberRepo.GetByID(ctx, member.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Member Name", updated.Name)
		assert.Equal(t, newUserID, *updated.UserID)
	})
	
	t.Run("non-existent member", func(t *testing.T) {
		member := createTestMember(group.ID)
		
		err := memberRepo.Update(ctx, member)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member not found")
	})
}

func TestPostgresMemberRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	ctx := context.Background()
	
	// Create a test group first
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	t.Run("successful deletion", func(t *testing.T) {
		member := createTestMember(group.ID)
		err := memberRepo.Create(ctx, member)
		require.NoError(t, err)
		
		err = memberRepo.Delete(ctx, member.ID)
		assert.NoError(t, err)
		
		// Verify the member is deleted
		_, err = memberRepo.GetByID(ctx, member.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member not found")
	})
	
	t.Run("non-existent member", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		
		err := memberRepo.Delete(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "member not found")
	})
}

func TestPostgresMemberRepository_ExistsByGroupAndUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	ctx := context.Background()
	
	// Create a test group first
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	userID := uuid.New().String()
	
	t.Run("existing membership", func(t *testing.T) {
		member := createTestMember(group.ID)
		member.UserID = &userID
		err := memberRepo.Create(ctx, member)
		require.NoError(t, err)
		
		exists, err := memberRepo.ExistsByGroupAndUser(ctx, group.ID, userID)
		require.NoError(t, err)
		assert.True(t, exists)
	})
	
	t.Run("non-existing membership", func(t *testing.T) {
		nonExistentUserID := uuid.New().String()
		
		exists, err := memberRepo.ExistsByGroupAndUser(ctx, group.ID, nonExistentUserID)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestPostgresMemberRepository_GetCreatorByGroupID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	ctx := context.Background()
	
	// Create a test group first
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	t.Run("group with creator", func(t *testing.T) {
		// Create creator member
		creator := createTestMember(group.ID)
		creator.IsCreator = true
		err := memberRepo.Create(ctx, creator)
		require.NoError(t, err)
		
		// Create regular member
		regularMember := createTestMember(group.ID)
		regularMember.IsCreator = false
		err = memberRepo.Create(ctx, regularMember)
		require.NoError(t, err)
		
		retrievedCreator, err := memberRepo.GetCreatorByGroupID(ctx, group.ID)
		require.NoError(t, err)
		assert.Equal(t, creator.ID, retrievedCreator.ID)
		assert.True(t, retrievedCreator.IsCreator)
	})
	
	t.Run("group without creator", func(t *testing.T) {
		emptyGroup := createTestGroup()
		err := groupRepo.Create(ctx, emptyGroup)
		require.NoError(t, err)
		
		_, err = memberRepo.GetCreatorByGroupID(ctx, emptyGroup.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "group creator not found")
	})
}