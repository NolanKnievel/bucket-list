package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"collaborative-bucket-list/internal/models"
)

// PostgresBucketItemRepository implements BucketItemRepository for PostgreSQL
type PostgresBucketItemRepository struct {
	db dbExecutor
}



// NewPostgresBucketItemRepository creates a new PostgreSQL bucket item repository
func NewPostgresBucketItemRepository(db dbExecutor) *PostgresBucketItemRepository {
	return &PostgresBucketItemRepository{db: db}
}

// Create creates a new bucket list item
func (r *PostgresBucketItemRepository) Create(ctx context.Context, item *models.BucketListItem) error {
	if err := item.IsValid(); err != nil {
		return fmt.Errorf("invalid bucket item data: %w", err)
	}

	item.Sanitize()

	query := `
		INSERT INTO bucket_items (id, group_id, title, description, completed, 
								 completed_by, completed_at, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.GroupID, item.Title, item.Description, item.Completed,
		item.CompletedBy, item.CompletedAt, item.CreatedBy, item.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create bucket item: %w", err)
	}

	return nil
}

// GetByID retrieves a bucket list item by its ID
func (r *PostgresBucketItemRepository) GetByID(ctx context.Context, id string) (*models.BucketListItem, error) {
	query := `
		SELECT id, group_id, title, description, completed, completed_by,
			   completed_at, created_by, created_at
		FROM bucket_items
		WHERE id = $1`

	var item models.BucketListItem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.GroupID, &item.Title, &item.Description, &item.Completed,
		&item.CompletedBy, &item.CompletedAt, &item.CreatedBy, &item.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bucket item not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get bucket item: %w", err)
	}

	return &item, nil
}

// GetByGroupID retrieves all items for a specific group
func (r *PostgresBucketItemRepository) GetByGroupID(ctx context.Context, groupID string) ([]models.BucketListItem, error) {
	query := `
		SELECT id, group_id, title, description, completed, completed_by,
			   completed_at, created_by, created_at
		FROM bucket_items
		WHERE group_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket items by group ID: %w", err)
	}
	defer rows.Close()

	var items []models.BucketListItem
	for rows.Next() {
		var item models.BucketListItem
		err := rows.Scan(&item.ID, &item.GroupID, &item.Title, &item.Description,
			&item.Completed, &item.CompletedBy, &item.CompletedAt, &item.CreatedBy, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bucket item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bucket items: %w", err)
	}

	return items, nil
}

// Update updates an existing bucket list item
func (r *PostgresBucketItemRepository) Update(ctx context.Context, item *models.BucketListItem) error {
	if err := item.IsValid(); err != nil {
		return fmt.Errorf("invalid bucket item data: %w", err)
	}

	item.Sanitize()

	query := `
		UPDATE bucket_items
		SET title = $2, description = $3, completed = $4, completed_by = $5, completed_at = $6
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		item.ID, item.Title, item.Description, item.Completed, item.CompletedBy, item.CompletedAt)
	if err != nil {
		return fmt.Errorf("failed to update bucket item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bucket item not found: %s", item.ID)
	}

	return nil
}

// Delete deletes a bucket list item by ID
func (r *PostgresBucketItemRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM bucket_items WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete bucket item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bucket item not found: %s", id)
	}

	return nil
}

// ToggleCompletion toggles the completion status of an item
func (r *PostgresBucketItemRepository) ToggleCompletion(ctx context.Context, itemID, memberID string, completed bool) error {
	var query string
	var args []interface{}

	if completed {
		// Mark as completed
		query = `
			UPDATE bucket_items
			SET completed = true, completed_by = $2, completed_at = $3
			WHERE id = $1`
		args = []interface{}{itemID, memberID, time.Now()}
	} else {
		// Mark as not completed
		query = `
			UPDATE bucket_items
			SET completed = false, completed_by = NULL, completed_at = NULL
			WHERE id = $1`
		args = []interface{}{itemID}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to toggle completion status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bucket item not found: %s", itemID)
	}

	return nil
}

// GetCompletionStats returns completion statistics for a group
func (r *PostgresBucketItemRepository) GetCompletionStats(ctx context.Context, groupID string) (total, completed int, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN completed = true THEN 1 END) as completed
		FROM bucket_items
		WHERE group_id = $1`

	err = r.db.QueryRowContext(ctx, query, groupID).Scan(&total, &completed)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get completion stats: %w", err)
	}

	return total, completed, nil
}