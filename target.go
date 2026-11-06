package core

import "context"

// SchemaManager provisions the target database.
//
// Recreate replaces the existing target database and creates a fresh
// schema. Implementations must handle database-specific lifecycle
// operations.
type SchemaManager interface {
	Recreate(ctx context.Context, schema Schema) error
}

// RecordWriter inserts a record into a target table.
//
// Insert returns generated column values in the returned Record,
// keyed by their target column names. Values explicitly supplied
// by the caller may also be returned.
//
// An error means the record was not successfully inserted.
type RecordWriter interface {
	Insert(ctx context.Context, table string, record Record) (Record, error)
}

// Transaction represents an atomic sequence of database operations.
//
// All operations performed through this interface belong to the
// same transaction. Commit makes them permanent. Rollback discards
// them.
//
// Rollback must be safe to call after Commit or a previous Rollback.
type Transaction interface {
	RecordWriter
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Target represents a database target.
type Target interface {
	SchemaManager
	Begin(ctx context.Context) (Transaction, error)
	Close() error
}
