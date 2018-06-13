// Package migration gives the structures and tools to handle versioned SQL database
// migration. In order to work a _migrations table is added to the database
package migration

import (
	"errors"
	"io/ioutil"
	"log"

	"github.com/almerlucke/go-utils/sql/database"
	"github.com/almerlucke/go-utils/sql/model"
	"github.com/almerlucke/go-utils/sql/types"
)

type (
	// Info contains models database meta information
	Info struct {
		ID            int64          `db:"id"`
		Version       string         `db:"version" sql:"override,VARCHAR(64)"`
		MigrationDate types.DateTime `db:"migration_date"`
	}

	// CustomMigrationFunc custom migration function to be run during migration
	CustomMigrationFunc func(queryer database.Queryer) error

	// Migration interface type
	Migration interface {
		Migrate(database.Queryer) error
	}

	// QueryMigration migrate by direct query
	QueryMigration struct {
		Query string
	}

	// ScriptMigration migrate by SQL script file (can contain only one SQL query)
	ScriptMigration struct {
		Script string
	}

	// CustomMigration migrate by calling a custom function
	CustomMigration struct {
		Func CustomMigrationFunc
	}

	// Version for grouping migrations
	Version struct {
		version    string
		migrations []Migration
	}
)

// Global migration tabler
var _migrationTable model.Tabler

// Initialize table
func init() {
	table, err := model.NewTable("_migration", &Info{})
	if err != nil {
		log.Fatalf("failed to create migration table %v", err)
	}

	_migrationTable = table
}

// Migrate migrate via direct query string
func (migration *QueryMigration) Migrate(queryer database.Queryer) error {
	_, err := queryer.Exec(migration.Query)
	return err
}

// Migrate migrate via SQL script
func (migration *ScriptMigration) Migrate(queryer database.Queryer) error {
	queryBytes, err := ioutil.ReadFile(migration.Script)
	if err != nil {
		return err
	}

	_, err = queryer.Exec(string(queryBytes))
	return err
}

// Migrate migrate via custom function
func (migration *CustomMigration) Migrate(queryer database.Queryer) error {
	return migration.Func(queryer)
}

// Migrate performs all migrations for a version
func (version *Version) Migrate(queryer database.Queryer) error {
	for _, migration := range version.migrations {
		err := migration.Migrate(queryer)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewQueryMigration create a new migration with a query
func NewQueryMigration(query string) Migration {
	return &QueryMigration{Query: query}
}

// NewScriptMigration create a new migration from a SQL script
func NewScriptMigration(script string) Migration {
	return &ScriptMigration{Script: script}
}

// NewCustomMigration create a new migration with a custom func
func NewCustomMigration(customFunc CustomMigrationFunc) Migration {
	return &CustomMigration{Func: customFunc}
}

// NewVersion create a new migration version
func NewVersion(version string, migrations []Migration) *Version {
	return &Version{version: version, migrations: migrations}
}

// Migrate database versions
func Migrate(queryer database.Queryer, currentVersion string, versions []*Version) error {
	// Create table if not exists
	_, err := queryer.Exec(_migrationTable.TableQuery())
	if err != nil {
		return err
	}

	// Get info row
	result, err := _migrationTable.Select("*").Run(queryer)
	if err != nil {
		return err
	}

	// Prepare info
	info := &Info{ID: 1, Version: "0", MigrationDate: types.NewDateTime()}
	rows := result.([]*Info)
	if len(rows) == 0 {
		_, err := _migrationTable.Insert([]interface{}{info}, queryer)
		if err != nil {
			return err
		}
	} else {
		info = rows[0]
	}

	// If current version is greater than database version we need to run migrations
	if currentVersion > info.Version {
		for _, migrationVersion := range versions {
			// We only perform migrations for versions up to info version and including current version
			if info.Version < migrationVersion.version && migrationVersion.version <= currentVersion {
				// Perform migration of the version
				migrationErr := migrationVersion.Migrate(queryer)
				if migrationErr != nil {
					return migrationErr
				}
			}
		}

		// Update info version
		info.Version = currentVersion
		info.MigrationDate = types.NewDateTime()

		_, err = _migrationTable.Update(info, queryer)
		if err != nil {
			return err
		}
	} else if currentVersion < info.Version {
		// The current code version is lacking behind the database version, this is not allowed
		return errors.New("database migration version is greater than current version")
	}

	return nil
}
