// Package sql defines DB, Querier and Time convenience methods and structures
package sql

import (
	"database/sql"

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
	db, err := sqlx.Open("mysql", config.ConnectionString())
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

// Transactional performs a given function wrapped inside a transaction
func (db *DB) Transactional(fn func(queryer Queryer) error) error {
	// Start transaction
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// Perform transactional function
	err = fn(tx)
	if err != nil {
		// Rollback all changes
		tx.Rollback()
		return err
	}

	// Commit changes
	return tx.Commit()
}
