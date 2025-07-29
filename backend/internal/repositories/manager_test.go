package repositories

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresRepositoryManager_Creation(t *testing.T) {
	t.Run("manager creation without database", func(t *testing.T) {
		// Test that we can create a manager instance without a real database connection
		// This tests the structure and interface compliance
		var db *sql.DB // nil database for structure testing
		
		// This should not panic and should create the manager
		manager := NewPostgresRepositoryManager(db)
		
		// Verify that all repository getters return non-nil interfaces
		assert.NotNil(t, manager.Groups())
		assert.NotNil(t, manager.Members())
		assert.NotNil(t, manager.BucketItems())
	})
}

func TestTransactionalRepositoryManager_Creation(t *testing.T) {
	t.Run("transactional manager creation without transaction", func(t *testing.T) {
		// Test that we can create a transactional manager instance
		var tx *sql.Tx // nil transaction for structure testing
		
		// This should not panic and should create the manager
		manager := NewTransactionalRepositoryManager(tx)
		
		// Verify that all repository getters return non-nil interfaces
		assert.NotNil(t, manager.Groups())
		assert.NotNil(t, manager.Members())
		assert.NotNil(t, manager.BucketItems())
	})
	
	t.Run("nested transactions not supported", func(t *testing.T) {
		var tx *sql.Tx
		manager := NewTransactionalRepositoryManager(tx)
		
		err := manager.WithTx(context.Background(), func(tx *sql.Tx) error {
			return nil
		})
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nested transactions are not supported")
	})
}

func TestRepositoryInterfaces(t *testing.T) {
	t.Run("repository manager implements RepositoryManager interface", func(t *testing.T) {
		var db *sql.DB
		manager := NewPostgresRepositoryManager(db)
		
		// Verify that the manager implements the RepositoryManager interface
		var _ RepositoryManager = manager
	})
	
	t.Run("transactional manager implements RepositoryManager interface", func(t *testing.T) {
		var tx *sql.Tx
		manager := NewTransactionalRepositoryManager(tx)
		
		// Verify that the transactional manager implements the RepositoryManager interface
		var _ RepositoryManager = manager
	})
	
	t.Run("repositories implement their respective interfaces", func(t *testing.T) {
		var db *sql.DB
		
		// Test that repository constructors return correct interface types
		groupRepo := NewPostgresGroupRepository(db)
		memberRepo := NewPostgresMemberRepository(db)
		bucketItemRepo := NewPostgresBucketItemRepository(db)
		
		// Verify interface compliance
		var _ GroupRepository = groupRepo
		var _ MemberRepository = memberRepo
		var _ BucketItemRepository = bucketItemRepo
	})
}