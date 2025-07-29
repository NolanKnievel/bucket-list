package repositories

import (
	"context"
	"database/sql"

	"collaborative-bucket-list/internal/models"
)

// GroupRepository defines the interface for group data operations
type GroupRepository interface {
	// Create creates a new group
	Create(ctx context.Context, group *models.Group) error
	
	// GetByID retrieves a group by its ID
	GetByID(ctx context.Context, id string) (*models.Group, error)
	
	// GetByUserID retrieves all groups created by a specific user
	GetByUserID(ctx context.Context, userID string) ([]models.Group, error)
	
	// Update updates an existing group
	Update(ctx context.Context, group *models.Group) error
	
	// Delete deletes a group by ID
	Delete(ctx context.Context, id string) error
	
	// GetWithDetails retrieves a group with all its members and items
	GetWithDetails(ctx context.Context, id string) (*models.GroupWithDetails, error)
	
	// GetSummariesByUserID retrieves group summaries for a user's dashboard
	GetSummariesByUserID(ctx context.Context, userID string) ([]models.GroupSummary, error)
}

// MemberRepository defines the interface for member data operations
type MemberRepository interface {
	// Create creates a new member
	Create(ctx context.Context, member *models.Member) error
	
	// GetByID retrieves a member by their ID
	GetByID(ctx context.Context, id string) (*models.Member, error)
	
	// GetByGroupID retrieves all members of a specific group
	GetByGroupID(ctx context.Context, groupID string) ([]models.Member, error)
	
	// GetByUserID retrieves all memberships for a specific user
	GetByUserID(ctx context.Context, userID string) ([]models.Member, error)
	
	// Update updates an existing member
	Update(ctx context.Context, member *models.Member) error
	
	// Delete deletes a member by ID
	Delete(ctx context.Context, id string) error
	
	// ExistsByGroupAndUser checks if a user is already a member of a group
	ExistsByGroupAndUser(ctx context.Context, groupID, userID string) (bool, error)
	
	// GetCreatorByGroupID retrieves the creator member of a group
	GetCreatorByGroupID(ctx context.Context, groupID string) (*models.Member, error)
}

// BucketItemRepository defines the interface for bucket list item data operations
type BucketItemRepository interface {
	// Create creates a new bucket list item
	Create(ctx context.Context, item *models.BucketListItem) error
	
	// GetByID retrieves a bucket list item by its ID
	GetByID(ctx context.Context, id string) (*models.BucketListItem, error)
	
	// GetByGroupID retrieves all items for a specific group
	GetByGroupID(ctx context.Context, groupID string) ([]models.BucketListItem, error)
	
	// Update updates an existing bucket list item
	Update(ctx context.Context, item *models.BucketListItem) error
	
	// Delete deletes a bucket list item by ID
	Delete(ctx context.Context, id string) error
	
	// ToggleCompletion toggles the completion status of an item
	ToggleCompletion(ctx context.Context, itemID, memberID string, completed bool) error
	
	// GetCompletionStats returns completion statistics for a group
	GetCompletionStats(ctx context.Context, groupID string) (total, completed int, err error)
}

// Repositories aggregates all repository interfaces
type Repositories struct {
	Groups      GroupRepository
	Members     MemberRepository
	BucketItems BucketItemRepository
}

// Transactional interface for operations that need database transactions
type Transactional interface {
	// WithTx executes a function within a database transaction
	WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error
}

// RepositoryManager manages repository instances and transactions
type RepositoryManager interface {
	Transactional
	Groups() GroupRepository
	Members() MemberRepository
	BucketItems() BucketItemRepository
}