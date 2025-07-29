package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"collaborative-bucket-list/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// Use environment variables or Docker defaults
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5432")
	user := getEnvOrDefault("TEST_DB_USER", "postgres")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "postgres")
	dbname := getEnvOrDefault("TEST_DB_NAME", "collaborative_bucket_list")
	sslmode := getEnvOrDefault("TEST_DB_SSL_MODE", "disable")
	
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("Skipping database tests: %v", err)
	}
	
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping database tests: database not available: %v", err)
	}
	
	// Create test database if it doesn't exist
	createTestDatabase(t, db, dbname)
	
	// Run migrations to ensure tables exist
	runTestMigrations(t, db)
	
	// Clean up tables before each test
	cleanupTables(t, db)
	
	return db
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// createTestDatabase creates the test database if it doesn't exist
func createTestDatabase(t *testing.T, db *sql.DB, dbname string) {
	// Check if database exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)", dbname).Scan(&exists)
	if err != nil {
		t.Logf("Warning: Could not check if database exists: %v", err)
		return
	}
	
	if !exists {
		t.Logf("Creating test database: %s", dbname)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			t.Logf("Warning: Could not create test database: %v", err)
		}
	}
}

// runTestMigrations runs the database migrations for tests
func runTestMigrations(t *testing.T, db *sql.DB) {
	// Enable UUID extension
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	if err != nil {
		t.Logf("Warning: Could not create uuid-ossp extension: %v", err)
	}
	
	// Create groups table
	createGroupsTable := `
		CREATE TABLE IF NOT EXISTS groups (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name TEXT NOT NULL,
			deadline TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by UUID NOT NULL
		)`
	_, err = db.Exec(createGroupsTable)
	require.NoError(t, err, "Failed to create groups table")
	
	// Create members table
	createMembersTable := `
		CREATE TABLE IF NOT EXISTS members (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
			user_id UUID,
			name TEXT NOT NULL,
			joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			is_creator BOOLEAN DEFAULT FALSE
		)`
	_, err = db.Exec(createMembersTable)
	require.NoError(t, err, "Failed to create members table")
	
	// Create bucket_items table
	createBucketItemsTable := `
		CREATE TABLE IF NOT EXISTS bucket_items (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			description TEXT,
			completed BOOLEAN DEFAULT FALSE,
			completed_by UUID REFERENCES members(id),
			completed_at TIMESTAMPTZ,
			created_by UUID NOT NULL REFERENCES members(id),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`
	_, err = db.Exec(createBucketItemsTable)
	require.NoError(t, err, "Failed to create bucket_items table")
}

// cleanupTables removes all data from test tables
func cleanupTables(t *testing.T, db *sql.DB) {
	tables := []string{"bucket_items", "members", "groups"}
	for _, table := range tables {
		_, err := db.Exec("DELETE FROM " + table)
		require.NoError(t, err, "Failed to clean up table: %s", table)
	}
}

// createTestGroup creates a test group for use in tests
func createTestGroup() *models.Group {
	return &models.Group{
		ID:        uuid.New().String(),
		Name:      "Test Group",
		Deadline:  nil,
		CreatedAt: time.Now(),
		CreatedBy: uuid.New().String(),
	}
}

func TestPostgresGroupRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewPostgresGroupRepository(db)
	ctx := context.Background()
	
	t.Run("successful creation", func(t *testing.T) {
		group := createTestGroup()
		
		err := repo.Create(ctx, group)
		assert.NoError(t, err)
		
		// Verify the group was created
		retrieved, err := repo.GetByID(ctx, group.ID)
		require.NoError(t, err)
		assert.Equal(t, group.ID, retrieved.ID)
		assert.Equal(t, group.Name, retrieved.Name)
		assert.Equal(t, group.CreatedBy, retrieved.CreatedBy)
	})
	
	t.Run("invalid group data", func(t *testing.T) {
		group := &models.Group{
			ID:        uuid.New().String(),
			Name:      "", // Invalid: empty name
			CreatedAt: time.Now(),
			CreatedBy: uuid.New().String(),
		}
		
		err := repo.Create(ctx, group)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid group data")
	})
	
	t.Run("duplicate ID", func(t *testing.T) {
		group1 := createTestGroup()
		group2 := createTestGroup()
		group2.ID = group1.ID // Same ID
		
		err := repo.Create(ctx, group1)
		require.NoError(t, err)
		
		err = repo.Create(ctx, group2)
		assert.Error(t, err)
	})
}

func TestPostgresGroupRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewPostgresGroupRepository(db)
	ctx := context.Background()
	
	t.Run("existing group", func(t *testing.T) {
		group := createTestGroup()
		err := repo.Create(ctx, group)
		require.NoError(t, err)
		
		retrieved, err := repo.GetByID(ctx, group.ID)
		require.NoError(t, err)
		assert.Equal(t, group.ID, retrieved.ID)
		assert.Equal(t, group.Name, retrieved.Name)
		assert.Equal(t, group.CreatedBy, retrieved.CreatedBy)
	})
	
	t.Run("non-existent group", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		
		_, err := repo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "group not found")
	})
}

func TestPostgresGroupRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewPostgresGroupRepository(db)
	ctx := context.Background()
	
	userID := uuid.New().String()
	
	t.Run("user with groups", func(t *testing.T) {
		// Create multiple groups for the user
		group1 := createTestGroup()
		group1.CreatedBy = userID
		group2 := createTestGroup()
		group2.CreatedBy = userID
		
		err := repo.Create(ctx, group1)
		require.NoError(t, err)
		err = repo.Create(ctx, group2)
		require.NoError(t, err)
		
		// Create a group for another user
		otherGroup := createTestGroup()
		otherGroup.CreatedBy = uuid.New().String()
		err = repo.Create(ctx, otherGroup)
		require.NoError(t, err)
		
		groups, err := repo.GetByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, groups, 2)
		
		// Verify all groups belong to the user
		for _, group := range groups {
			assert.Equal(t, userID, group.CreatedBy)
		}
	})
	
	t.Run("user with no groups", func(t *testing.T) {
		nonExistentUserID := uuid.New().String()
		
		groups, err := repo.GetByUserID(ctx, nonExistentUserID)
		require.NoError(t, err)
		assert.Empty(t, groups)
	})
}

func TestPostgresGroupRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewPostgresGroupRepository(db)
	ctx := context.Background()
	
	t.Run("successful update", func(t *testing.T) {
		group := createTestGroup()
		err := repo.Create(ctx, group)
		require.NoError(t, err)
		
		// Update the group
		group.Name = "Updated Group Name"
		deadline := time.Now().Add(24 * time.Hour)
		group.Deadline = &deadline
		
		err = repo.Update(ctx, group)
		assert.NoError(t, err)
		
		// Verify the update
		updated, err := repo.GetByID(ctx, group.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Group Name", updated.Name)
		assert.NotNil(t, updated.Deadline)
	})
	
	t.Run("non-existent group", func(t *testing.T) {
		group := createTestGroup()
		
		err := repo.Update(ctx, group)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "group not found")
	})
}

func TestPostgresGroupRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewPostgresGroupRepository(db)
	ctx := context.Background()
	
	t.Run("successful deletion", func(t *testing.T) {
		group := createTestGroup()
		err := repo.Create(ctx, group)
		require.NoError(t, err)
		
		err = repo.Delete(ctx, group.ID)
		assert.NoError(t, err)
		
		// Verify the group is deleted
		_, err = repo.GetByID(ctx, group.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "group not found")
	})
	
	t.Run("non-existent group", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		
		err := repo.Delete(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "group not found")
	})
}

func TestPostgresGroupRepository_GetWithDetails(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	itemRepo := NewPostgresBucketItemRepository(db)
	ctx := context.Background()
	
	t.Run("group with members and items", func(t *testing.T) {
		// Create a group
		group := createTestGroup()
		err := repo.Create(ctx, group)
		require.NoError(t, err)
		
		// Create members
		member1 := &models.Member{
			ID:        uuid.New().String(),
			GroupID:   group.ID,
			Name:      "Member 1",
			JoinedAt:  time.Now(),
			IsCreator: true,
		}
		member2 := &models.Member{
			ID:        uuid.New().String(),
			GroupID:   group.ID,
			Name:      "Member 2",
			JoinedAt:  time.Now(),
			IsCreator: false,
		}
		
		err = memberRepo.Create(ctx, member1)
		require.NoError(t, err)
		err = memberRepo.Create(ctx, member2)
		require.NoError(t, err)
		
		// Create bucket items
		item1 := &models.BucketListItem{
			ID:        uuid.New().String(),
			GroupID:   group.ID,
			Title:     "Item 1",
			CreatedBy: member1.ID,
			CreatedAt: time.Now(),
		}
		item2 := &models.BucketListItem{
			ID:        uuid.New().String(),
			GroupID:   group.ID,
			Title:     "Item 2",
			CreatedBy: member2.ID,
			CreatedAt: time.Now(),
		}
		
		err = itemRepo.Create(ctx, item1)
		require.NoError(t, err)
		err = itemRepo.Create(ctx, item2)
		require.NoError(t, err)
		
		// Get group with details
		details, err := repo.GetWithDetails(ctx, group.ID)
		require.NoError(t, err)
		
		assert.Equal(t, group.ID, details.ID)
		assert.Len(t, details.Members, 2)
		assert.Len(t, details.Items, 2)
	})
	
	t.Run("non-existent group", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		
		_, err := repo.GetWithDetails(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "group not found")
	})
}

func TestPostgresGroupRepository_GetSummariesByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewPostgresGroupRepository(db)
	memberRepo := NewPostgresMemberRepository(db)
	itemRepo := NewPostgresBucketItemRepository(db)
	ctx := context.Background()
	
	userID := uuid.New().String()
	
	t.Run("user with group summaries", func(t *testing.T) {
		// Create a group
		group := createTestGroup()
		group.CreatedBy = userID
		err := repo.Create(ctx, group)
		require.NoError(t, err)
		
		// Create members
		member := &models.Member{
			ID:        uuid.New().String(),
			GroupID:   group.ID,
			UserID:    &userID,
			Name:      "Test Member",
			JoinedAt:  time.Now(),
			IsCreator: true,
		}
		err = memberRepo.Create(ctx, member)
		require.NoError(t, err)
		
		// Create bucket items
		item1 := &models.BucketListItem{
			ID:        uuid.New().String(),
			GroupID:   group.ID,
			Title:     "Item 1",
			Completed: false,
			CreatedBy: member.ID,
			CreatedAt: time.Now(),
		}
		item2 := &models.BucketListItem{
			ID:        uuid.New().String(),
			GroupID:   group.ID,
			Title:     "Item 2",
			Completed: true,
			CreatedBy: member.ID,
			CreatedAt: time.Now(),
		}
		
		err = itemRepo.Create(ctx, item1)
		require.NoError(t, err)
		err = itemRepo.Create(ctx, item2)
		require.NoError(t, err)
		
		// Get summaries
		summaries, err := repo.GetSummariesByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, summaries, 1)
		
		summary := summaries[0]
		assert.Equal(t, group.ID, summary.ID)
		assert.Equal(t, 1, summary.MemberCount)
		assert.Equal(t, 2, summary.ItemCount)
		assert.Equal(t, 1, summary.CompletedCount)
		assert.Equal(t, 50.0, summary.ProgressPercent)
	})
}