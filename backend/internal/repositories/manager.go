package repositories

import (
	"context"
	"database/sql"
	"fmt"
)

// PostgresRepositoryManager implements RepositoryManager for PostgreSQL
type PostgresRepositoryManager struct {
	db          *sql.DB
	groups      GroupRepository
	members     MemberRepository
	bucketItems BucketItemRepository
}

// NewPostgresRepositoryManager creates a new PostgreSQL repository manager
func NewPostgresRepositoryManager(db *sql.DB) *PostgresRepositoryManager {
	return &PostgresRepositoryManager{
		db:          db,
		groups:      NewPostgresGroupRepository(db),
		members:     NewPostgresMemberRepository(db),
		bucketItems: NewPostgresBucketItemRepository(db),
	}
}

// Groups returns the group repository
func (m *PostgresRepositoryManager) Groups() GroupRepository {
	return m.groups
}

// Members returns the member repository
func (m *PostgresRepositoryManager) Members() MemberRepository {
	return m.members
}

// BucketItems returns the bucket item repository
func (m *PostgresRepositoryManager) BucketItems() BucketItemRepository {
	return m.bucketItems
}

// WithTx executes a function within a database transaction
func (m *PostgresRepositoryManager) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TransactionalRepositoryManager provides transactional repository access
type TransactionalRepositoryManager struct {
	tx          *sql.Tx
	groups      GroupRepository
	members     MemberRepository
	bucketItems BucketItemRepository
}

// NewTransactionalRepositoryManager creates repository manager for use within a transaction
func NewTransactionalRepositoryManager(tx *sql.Tx) *TransactionalRepositoryManager {
	return &TransactionalRepositoryManager{
		tx:          tx,
		groups:      NewPostgresGroupRepository(tx),
		members:     NewPostgresMemberRepository(tx),
		bucketItems: NewPostgresBucketItemRepository(tx),
	}
}

// Groups returns the group repository
func (m *TransactionalRepositoryManager) Groups() GroupRepository {
	return m.groups
}

// Members returns the member repository
func (m *TransactionalRepositoryManager) Members() MemberRepository {
	return m.members
}

// BucketItems returns the bucket item repository
func (m *TransactionalRepositoryManager) BucketItems() BucketItemRepository {
	return m.bucketItems
}

// WithTx is not supported within a transactional manager
func (m *TransactionalRepositoryManager) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return fmt.Errorf("nested transactions are not supported")
}

// Helper function to create repositories that work with both *sql.DB and *sql.Tx
type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}