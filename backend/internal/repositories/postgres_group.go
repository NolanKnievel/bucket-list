package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"collaborative-bucket-list/internal/models"
)

// PostgresGroupRepository implements GroupRepository for PostgreSQL
type PostgresGroupRepository struct {
	db dbExecutor
}



// NewPostgresGroupRepository creates a new PostgreSQL group repository
func NewPostgresGroupRepository(db dbExecutor) *PostgresGroupRepository {
	return &PostgresGroupRepository{db: db}
}

// Create creates a new group
func (r *PostgresGroupRepository) Create(ctx context.Context, group *models.Group) error {
	if err := group.IsValid(); err != nil {
		return fmt.Errorf("invalid group data: %w", err)
	}

	group.Sanitize()

	query := `
		INSERT INTO groups (id, name, deadline, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		group.ID, group.Name, group.Deadline, group.CreatedAt, group.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	return nil
}

// GetByID retrieves a group by its ID
func (r *PostgresGroupRepository) GetByID(ctx context.Context, id string) (*models.Group, error) {
	query := `
		SELECT id, name, deadline, created_at, created_by
		FROM groups
		WHERE id = $1`

	var group models.Group
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&group.ID, &group.Name, &group.Deadline, &group.CreatedAt, &group.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("group not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return &group, nil
}

// GetByUserID retrieves all groups created by a specific user
func (r *PostgresGroupRepository) GetByUserID(ctx context.Context, userID string) ([]models.Group, error) {
	query := `
		SELECT id, name, deadline, created_at, created_by
		FROM groups
		WHERE created_by = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups by user ID: %w", err)
	}
	defer rows.Close()

	var groups []models.Group
	for rows.Next() {
		var group models.Group
		err := rows.Scan(&group.ID, &group.Name, &group.Deadline, &group.CreatedAt, &group.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating groups: %w", err)
	}

	return groups, nil
}

// Update updates an existing group
func (r *PostgresGroupRepository) Update(ctx context.Context, group *models.Group) error {
	if err := group.IsValid(); err != nil {
		return fmt.Errorf("invalid group data: %w", err)
	}

	group.Sanitize()

	query := `
		UPDATE groups
		SET name = $2, deadline = $3
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, group.ID, group.Name, group.Deadline)
	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("group not found: %s", group.ID)
	}

	return nil
}

// Delete deletes a group by ID
func (r *PostgresGroupRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM groups WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("group not found: %s", id)
	}

	return nil
}

// GetWithDetails retrieves a group with all its members and items
func (r *PostgresGroupRepository) GetWithDetails(ctx context.Context, id string) (*models.GroupWithDetails, error) {
	// First get the group
	group, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get members
	membersQuery := `
		SELECT id, group_id, user_id, name, joined_at, is_creator
		FROM members
		WHERE group_id = $1
		ORDER BY joined_at ASC`

	memberRows, err := r.db.QueryContext(ctx, membersQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer memberRows.Close()

	var members []models.Member
	for memberRows.Next() {
		var member models.Member
		err := memberRows.Scan(&member.ID, &member.GroupID, &member.UserID,
			&member.Name, &member.JoinedAt, &member.IsCreator)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, member)
	}

	if err = memberRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating members: %w", err)
	}

	// Get bucket list items
	itemsQuery := `
		SELECT id, group_id, title, description, completed, completed_by,
			   completed_at, created_by, created_at
		FROM bucket_items
		WHERE group_id = $1
		ORDER BY created_at DESC`

	itemRows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket items: %w", err)
	}
	defer itemRows.Close()

	var items []models.BucketListItem
	for itemRows.Next() {
		var item models.BucketListItem
		err := itemRows.Scan(&item.ID, &item.GroupID, &item.Title, &item.Description,
			&item.Completed, &item.CompletedBy, &item.CompletedAt, &item.CreatedBy, &item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bucket item: %w", err)
		}
		items = append(items, item)
	}

	if err = itemRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bucket items: %w", err)
	}

	return &models.GroupWithDetails{
		Group:   *group,
		Members: members,
		Items:   items,
	}, nil
}

// GetSummariesByUserID retrieves group summaries for a user's dashboard
func (r *PostgresGroupRepository) GetSummariesByUserID(ctx context.Context, userID string) ([]models.GroupSummary, error) {
	query := `
		SELECT 
			g.id, g.name, g.deadline, g.created_at, g.created_by,
			COUNT(DISTINCT m.id) as member_count,
			COUNT(DISTINCT bi.id) as item_count,
			COUNT(DISTINCT CASE WHEN bi.completed = true THEN bi.id END) as completed_count
		FROM groups g
		LEFT JOIN members m ON g.id = m.group_id
		LEFT JOIN bucket_items bi ON g.id = bi.group_id
		WHERE g.created_by = $1 OR g.id IN (
			SELECT group_id FROM members WHERE user_id = $1
		)
		GROUP BY g.id, g.name, g.deadline, g.created_at, g.created_by
		ORDER BY g.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group summaries: %w", err)
	}
	defer rows.Close()

	var summaries []models.GroupSummary
	for rows.Next() {
		var summary models.GroupSummary
		var memberCount, itemCount, completedCount int

		err := rows.Scan(
			&summary.ID, &summary.Name, &summary.Deadline, &summary.CreatedAt, &summary.CreatedBy,
			&memberCount, &itemCount, &completedCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group summary: %w", err)
		}

		summary.MemberCount = memberCount
		summary.ItemCount = itemCount
		summary.CompletedCount = completedCount

		// Calculate progress percentage
		if itemCount > 0 {
			summary.ProgressPercent = float64(completedCount) / float64(itemCount) * 100
		} else {
			summary.ProgressPercent = 0
		}

		summaries = append(summaries, summary)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group summaries: %w", err)
	}

	return summaries, nil
}