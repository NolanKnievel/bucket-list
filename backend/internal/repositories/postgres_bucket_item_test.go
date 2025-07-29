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

// createTestBucketItem creates a test bucket item for use in tests
func createTestBucketItem(groupID, memberID string) *models.BucketListItem {
	description := "Test item description"
	return &models.BucketListItem{
		ID:          uuid.New().String(),
		GroupID:     groupID,
		Title:       "Test Bucket Item",
		Description: &description,
		Completed:   false,
		CreatedBy:   memberID,
		CreatedAt:   time.Now(),
	}
}

func TestPostgresBucketItemRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	itemRepo := NewPostgresBucketItemRepository(db)
	ctx := context.Background()
	
	// Create test group and member
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	member := createTestMember(group.ID)
	err = memberRepo.Create(ctx, member)
	require.NoError(t, err)
	
	t.Run("successful creation", func(t *testing.T) {
		item := createTestBucketItem(group.ID, member.ID)
		
		err := itemRepo.Create(ctx, item)
		assert.NoError(t, err)
		
		// Verify the item was created
		retrieved, err := itemRepo.GetByID(ctx, item.ID)
		require.NoError(t, err)
		assert.Equal(t, item.ID, retrieved.ID)
		assert.Equal(t, item.Title, retrieved.Title)
		assert.Equal(t, *item.Description, *retrieved.Description)
		assert.Equal(t, item.GroupID, retrieved.GroupID)
		assert.Equal(t, item.CreatedBy, retrieved.CreatedBy)
		assert.False(t, retrieved.Completed)
	})
	
	t.Run("item without description", func(t *testing.T) {
		item := createTestBucketItem(group.ID, member.ID)
		item.Description = nil
		
		err := itemRepo.Create(ctx, item)
		assert.NoError(t, err)
		
		retrieved, err := itemRepo.GetByID(ctx, item.ID)
		require.NoError(t, err)
		assert.Nil(t, retrieved.Description)
	})
	
	t.Run("invalid item data", func(t *testing.T) {
		item := &models.BucketListItem{
			ID:        uuid.New().String(),
			GroupID:   group.ID,
			Title:     "", // Invalid: empty title
			CreatedBy: member.ID,
			CreatedAt: time.Now(),
		}
		
		err := itemRepo.Create(ctx, item)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid bucket item data")
	})
}

func TestPostgresBucketItemRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	itemRepo := NewPostgresBucketItemRepository(db)
	ctx := context.Background()
	
	// Create test group and member
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	member := createTestMember(group.ID)
	err = memberRepo.Create(ctx, member)
	require.NoError(t, err)
	
	t.Run("existing item", func(t *testing.T) {
		item := createTestBucketItem(group.ID, member.ID)
		err := itemRepo.Create(ctx, item)
		require.NoError(t, err)
		
		retrieved, err := itemRepo.GetByID(ctx, item.ID)
		require.NoError(t, err)
		assert.Equal(t, item.ID, retrieved.ID)
		assert.Equal(t, item.Title, retrieved.Title)
		assert.Equal(t, item.GroupID, retrieved.GroupID)
	})
	
	t.Run("non-existent item", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		
		_, err := itemRepo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bucket item not found")
	})
}

func TestPostgresBucketItemRepository_GetByGroupID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	itemRepo := NewPostgresBucketItemRepository(db)
	ctx := context.Background()
	
	// Create test groups and members
	group1 := createTestGroup()
	group2 := createTestGroup()
	err := groupRepo.Create(ctx, group1)
	require.NoError(t, err)
	err = groupRepo.Create(ctx, group2)
	require.NoError(t, err)
	
	member1 := createTestMember(group1.ID)
	member2 := createTestMember(group2.ID)
	err = memberRepo.Create(ctx, member1)
	require.NoError(t, err)
	err = memberRepo.Create(ctx, member2)
	require.NoError(t, err)
	
	t.Run("group with items", func(t *testing.T) {
		// Create items for group1
		item1 := createTestBucketItem(group1.ID, member1.ID)
		item1.CreatedAt = time.Now().Add(-2 * time.Hour) // Older item
		item2 := createTestBucketItem(group1.ID, member1.ID)
		item2.CreatedAt = time.Now().Add(-1 * time.Hour) // Newer item
		
		err := itemRepo.Create(ctx, item1)
		require.NoError(t, err)
		err = itemRepo.Create(ctx, item2)
		require.NoError(t, err)
		
		// Create an item for group2
		item3 := createTestBucketItem(group2.ID, member2.ID)
		err = itemRepo.Create(ctx, item3)
		require.NoError(t, err)
		
		items, err := itemRepo.GetByGroupID(ctx, group1.ID)
		require.NoError(t, err)
		assert.Len(t, items, 2)
		
		// Verify all items belong to group1
		for _, item := range items {
			assert.Equal(t, group1.ID, item.GroupID)
		}
		
		// Verify ordering (created_at DESC - newest first)
		assert.True(t, items[0].CreatedAt.After(items[1].CreatedAt))
	})
	
	t.Run("group with no items", func(t *testing.T) {
		emptyGroup := createTestGroup()
		err := groupRepo.Create(ctx, emptyGroup)
		require.NoError(t, err)
		
		items, err := itemRepo.GetByGroupID(ctx, emptyGroup.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
	})
}

func TestPostgresBucketItemRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	itemRepo := NewPostgresBucketItemRepository(db)
	ctx := context.Background()
	
	// Create test group and member
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	member := createTestMember(group.ID)
	err = memberRepo.Create(ctx, member)
	require.NoError(t, err)
	
	t.Run("successful update", func(t *testing.T) {
		item := createTestBucketItem(group.ID, member.ID)
		err := itemRepo.Create(ctx, item)
		require.NoError(t, err)
		
		// Update the item
		item.Title = "Updated Item Title"
		newDescription := "Updated description"
		item.Description = &newDescription
		item.Completed = true
		completedAt := time.Now()
		item.CompletedAt = &completedAt
		item.CompletedBy = &member.ID
		
		err = itemRepo.Update(ctx, item)
		assert.NoError(t, err)
		
		// Verify the update
		updated, err := itemRepo.GetByID(ctx, item.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Item Title", updated.Title)
		assert.Equal(t, "Updated description", *updated.Description)
		assert.True(t, updated.Completed)
		assert.NotNil(t, updated.CompletedAt)
		assert.Equal(t, member.ID, *updated.CompletedBy)
	})
	
	t.Run("non-existent item", func(t *testing.T) {
		item := createTestBucketItem(group.ID, member.ID)
		
		err := itemRepo.Update(ctx, item)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bucket item not found")
	})
}

func TestPostgresBucketItemRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	itemRepo := NewPostgresBucketItemRepository(db)
	ctx := context.Background()
	
	// Create test group and member
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	member := createTestMember(group.ID)
	err = memberRepo.Create(ctx, member)
	require.NoError(t, err)
	
	t.Run("successful deletion", func(t *testing.T) {
		item := createTestBucketItem(group.ID, member.ID)
		err := itemRepo.Create(ctx, item)
		require.NoError(t, err)
		
		err = itemRepo.Delete(ctx, item.ID)
		assert.NoError(t, err)
		
		// Verify the item is deleted
		_, err = itemRepo.GetByID(ctx, item.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bucket item not found")
	})
	
	t.Run("non-existent item", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		
		err := itemRepo.Delete(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bucket item not found")
	})
}

func TestPostgresBucketItemRepository_ToggleCompletion(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	itemRepo := NewPostgresBucketItemRepository(db)
	ctx := context.Background()
	
	// Create test group and member
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	member := createTestMember(group.ID)
	err = memberRepo.Create(ctx, member)
	require.NoError(t, err)
	
	t.Run("mark as completed", func(t *testing.T) {
		item := createTestBucketItem(group.ID, member.ID)
		item.Completed = false
		err := itemRepo.Create(ctx, item)
		require.NoError(t, err)
		
		err = itemRepo.ToggleCompletion(ctx, item.ID, member.ID, true)
		assert.NoError(t, err)
		
		// Verify the item is marked as completed
		updated, err := itemRepo.GetByID(ctx, item.ID)
		require.NoError(t, err)
		assert.True(t, updated.Completed)
		assert.Equal(t, member.ID, *updated.CompletedBy)
		assert.NotNil(t, updated.CompletedAt)
	})
	
	t.Run("mark as not completed", func(t *testing.T) {
		item := createTestBucketItem(group.ID, member.ID)
		item.Completed = true
		completedAt := time.Now()
		item.CompletedAt = &completedAt
		item.CompletedBy = &member.ID
		err := itemRepo.Create(ctx, item)
		require.NoError(t, err)
		
		err = itemRepo.ToggleCompletion(ctx, item.ID, member.ID, false)
		assert.NoError(t, err)
		
		// Verify the item is marked as not completed
		updated, err := itemRepo.GetByID(ctx, item.ID)
		require.NoError(t, err)
		assert.False(t, updated.Completed)
		assert.Nil(t, updated.CompletedBy)
		assert.Nil(t, updated.CompletedAt)
	})
	
	t.Run("non-existent item", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		
		err := itemRepo.ToggleCompletion(ctx, nonExistentID, member.ID, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bucket item not found")
	})
}

func TestPostgresBucketItemRepository_GetCompletionStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	groupRepo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	itemRepo := NewPostgresBucketItemRepository(db)
	ctx := context.Background()
	
	// Create test group and member
	group := createTestGroup()
	err := groupRepo.Create(ctx, group)
	require.NoError(t, err)
	
	member := createTestMember(group.ID)
	err = memberRepo.Create(ctx, member)
	require.NoError(t, err)
	
	t.Run("group with mixed completion status", func(t *testing.T) {
		// Create completed items
		completedItem1 := createTestBucketItem(group.ID, member.ID)
		completedItem1.Completed = true
		completedItem2 := createTestBucketItem(group.ID, member.ID)
		completedItem2.Completed = true
		
		// Create incomplete items
		incompleteItem1 := createTestBucketItem(group.ID, member.ID)
		incompleteItem1.Completed = false
		incompleteItem2 := createTestBucketItem(group.ID, member.ID)
		incompleteItem2.Completed = false
		incompleteItem3 := createTestBucketItem(group.ID, member.ID)
		incompleteItem3.Completed = false
		
		items := []*models.BucketListItem{
			completedItem1, completedItem2, incompleteItem1, incompleteItem2, incompleteItem3,
		}
		
		for _, item := range items {
			err := itemRepo.Create(ctx, item)
			require.NoError(t, err)
		}
		
		total, completed, err := itemRepo.GetCompletionStats(ctx, group.ID)
		require.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Equal(t, 2, completed)
	})
	
	t.Run("group with no items", func(t *testing.T) {
		emptyGroup := createTestGroup()
		err := groupRepo.Create(ctx, emptyGroup)
		require.NoError(t, err)
		
		total, completed, err := itemRepo.GetCompletionStats(ctx, emptyGroup.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Equal(t, 0, completed)
	})
	
	t.Run("group with all completed items", func(t *testing.T) {
		allCompletedGroup := createTestGroup()
		err := groupRepo.Create(ctx, allCompletedGroup)
		require.NoError(t, err)
		
		allCompletedMember := createTestMember(allCompletedGroup.ID)
		err = memberRepo.Create(ctx, allCompletedMember)
		require.NoError(t, err)
		
		// Create only completed items
		for i := 0; i < 3; i++ {
			item := createTestBucketItem(allCompletedGroup.ID, allCompletedMember.ID)
			item.Completed = true
			err := itemRepo.Create(ctx, item)
			require.NoError(t, err)
		}
		
		total, completed, err := itemRepo.GetCompletionStats(ctx, allCompletedGroup.ID)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Equal(t, 3, completed)
	})
}