package utils

import (
	"github.com/almerlucke/go-utils/sql/database"
	"github.com/almerlucke/go-utils/sql/migration"
	"github.com/almerlucke/go-utils/sql/model"
)

// NewDatabase with configuration, version and migrations, and finally a variable number of tables to create
func NewDatabase(config *database.Configuration, version string, migrations []*migration.Version, tables ...model.Tabler) (*database.DB, error) {
	// Create an open database
	db, err := database.New(config)
	if err != nil {
		return nil, err
	}

	// Create tables if not exist
	for _, table := range tables {
		_, err = db.Exec(table.TableQuery())
		if err != nil {
			return nil, err
		}
	}

	// Perform migrations if necessary
	err = migration.Migrate(db, version, migrations)
	if err != nil {
		return nil, err
	}

	return db, nil
}
