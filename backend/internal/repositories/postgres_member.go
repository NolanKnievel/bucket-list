package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"collaborative-bucket-list/internal/models"
)



// PostgresMemberRepository implements MemberRepository for PostgreSQL
type PostgresMemberRepository struct {
	db dbExecutor
}

// NewPostgresMemberRepository creates a new PostgreSQL member repository
func NewPostgresMemberRepository(db dbExecutor) *PostgresMemberRepository {
	return &PostgresMemberRepository{db: db}
}

// Create creates a new member
func (r *PostgresMemberRepository) Create(ctx context.Context, member *models.Member) error {
	if err := member.IsValid(); err != nil {
		return fmt.Errorf("invalid member data: %w", err)
	}

	member.Sanitize()

	query := `
		INSERT INTO members (id, group_id, user_id, name, joined_at, is_creator)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		member.ID, member.GroupID, member.UserID, member.Name, member.JoinedAt, member.IsCreator)
	if err != nil {
		return fmt.Errorf("failed to create member: %w", err)
	}

	return nil
}

// GetByID retrieves a member by their ID
func (r *PostgresMemberRepository) GetByID(ctx context.Context, id string) (*models.Member, error) {
	query := `
		SELECT id, group_id, user_id, name, joined_at, is_creator
		FROM members
		WHERE id = $1`

	var member models.Member
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&member.ID, &member.GroupID, &member.UserID, &member.Name, &member.JoinedAt, &member.IsCreator)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("member not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	return &member, nil
}

// GetByGroupID retrieves all members of a specific group
func (r *PostgresMemberRepository) GetByGroupID(ctx context.Context, groupID string) ([]models.Member, error) {
	query := `
		SELECT id, group_id, user_id, name, joined_at, is_creator
		FROM members
		WHERE group_id = $1
		ORDER BY joined_at ASC`

	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members by group ID: %w", err)
	}
	defer rows.Close()

	var members []models.Member
	for rows.Next() {
		var member models.Member
		err := rows.Scan(&member.ID, &member.GroupID, &member.UserID,
			&member.Name, &member.JoinedAt, &member.IsCreator)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating members: %w", err)
	}

	return members, nil
}

// GetByUserID retrieves all memberships for a specific user
func (r *PostgresMemberRepository) GetByUserID(ctx context.Context, userID string) ([]models.Member, error) {
	query := `
		SELECT id, group_id, user_id, name, joined_at, is_creator
		FROM members
		WHERE user_id = $1
		ORDER BY joined_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members by user ID: %w", err)
	}
	defer rows.Close()

	var members []models.Member
	for rows.Next() {
		var member models.Member
		err := rows.Scan(&member.ID, &member.GroupID, &member.UserID,
			&member.Name, &member.JoinedAt, &member.IsCreator)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating members: %w", err)
	}

	return members, nil
}

// Update updates an existing member
func (r *PostgresMemberRepository) Update(ctx context.Context, member *models.Member) error {
	if err := member.IsValid(); err != nil {
		return fmt.Errorf("invalid member data: %w", err)
	}

	member.Sanitize()

	query := `
		UPDATE members
		SET name = $2, user_id = $3
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, member.ID, member.Name, member.UserID)
	if err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found: %s", member.ID)
	}

	return nil
}

// Delete deletes a member by ID
func (r *PostgresMemberRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM members WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("member not found: %s", id)
	}

	return nil
}

// ExistsByGroupAndUser checks if a user is already a member of a group
func (r *PostgresMemberRepository) ExistsByGroupAndUser(ctx context.Context, groupID, userID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM members
		WHERE group_id = $1 AND user_id = $2`

	var count int
	err := r.db.QueryRowContext(ctx, query, groupID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check member existence: %w", err)
	}

	return count > 0, nil
}

// GetCreatorByGroupID retrieves the creator member of a group
func (r *PostgresMemberRepository) GetCreatorByGroupID(ctx context.Context, groupID string) (*models.Member, error) {
	query := `
		SELECT id, group_id, user_id, name, joined_at, is_creator
		FROM members
		WHERE group_id = $1 AND is_creator = true`

	var member models.Member
	err := r.db.QueryRowContext(ctx, query, groupID).Scan(
		&member.ID, &member.GroupID, &member.UserID, &member.Name, &member.JoinedAt, &member.IsCreator)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("group creator not found for group: %s", groupID)
		}
		return nil, fmt.Errorf("failed to get group creator: %w", err)
	}

	return &member, nil
}