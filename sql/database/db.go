package database

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DB wrapper around *sqlx.DB
type DB struct {
	*sqlx.DB
}

// Queryer is an interface to abstract Tx or DB
type Queryer interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// New database connection
func New(config *Configuration) (*DB, error) {
	db, err := sqlx.Open(config.SQLType, config.ConnectionString())
	if err != nil {
		return nil, err
	}

	// Ping the DB first
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Following methods can be used to tweak the connection pooling
	// db.SetConnMaxLifetime
	// db.SetMaxIdleConns
	// db.SetMaxOpenConns

	return &DB{DB: db}, nil
}

// Transactional performs a given function wrapped inside a transaction, if the function
// returns false or an error we perform a rollback
func (db *DB) Transactional(fn func(queryer Queryer) (bool, error)) error {
	// Start transaction
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// Perform transactional function
	commit, err := fn(tx)
	if err != nil {
		// Try to rollback all changes after an error
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return fmt.Errorf("rolback error: %v - when trying to rollback from error: %v", rollbackErr, err)
		}

		return err
	}

	if !commit {
		// Try to rollback all changes
		return tx.Rollback()
	}

	// Commit changes
	return tx.Commit()
}
