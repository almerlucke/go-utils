// Package migration gives the structures and tools to handle versioned SQL database
// migration. In order to work a _migrations table is added to the database
package migration

import (
	"errors"
	"io/ioutil"

	"github.com/almerlucke/go-utils/sql"
)

const (
	// MigrationTableCreateQuery query for creating migration table
	MigrationTableCreateQuery = `
		CREATE TABLE _migration (
			id int(11) unsigned NOT NULL AUTO_INCREMENT,
			version varchar(32) DEFAULT '0',
			migration_date datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	// MigrationEntryCreateQuery query to create initial migration info row
	MigrationEntryCreateQuery = `INSERT INTO _migration () VALUES ()`
)

type (
	// Info contains models database meta information
	Info struct {
		ID            int64        `db:"id"`
		Version       string       `db:"version"`
		MigrationDate sql.DateTime `db:"migration_date"`
	}

	// CustomMigrationFunc custom migration function to be run during migration
	CustomMigrationFunc func(queryer sql.Queryer) error

	// Migration interface type
	Migration interface {
		Migrate(sql.Queryer) error
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

// Migrate migrate via direct query string
func (migration *QueryMigration) Migrate(queryer sql.Queryer) error {
	_, err := queryer.Exec(migration.Query)
	return err
}

// Migrate migrate via SQL script
func (migration *ScriptMigration) Migrate(queryer sql.Queryer) error {
	queryBytes, err := ioutil.ReadFile(migration.Script)
	if err != nil {
		return err
	}

	_, err = queryer.Exec(string(queryBytes))
	return err
}

// Migrate migrate via custom function
func (migration *CustomMigration) Migrate(queryer sql.Queryer) error {
	return migration.Func(queryer)
}

// Migrate performs all migrations for a version
func (version *Version) Migrate(queryer sql.Queryer) error {
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

// GetMigrationInfo get migration info including database version for migrations
func GetMigrationInfo(queryer sql.Queryer) (*Info, error) {
	info := &Info{}

	err := queryer.Get(info, "SELECT * FROM _migration")
	if err != nil {
		// Migration info table does not exist yet, create it
		_, err = queryer.Exec(MigrationTableCreateQuery)
		if err != nil {
			return nil, err
		}
		// Insert initial migration info table
		_, err = queryer.Exec(MigrationEntryCreateQuery)
		if err != nil {
			return nil, err
		}
		// Get migration info
		err = queryer.Get(info, "SELECT * FROM _migration")
		if err != nil {
			return nil, err
		}
	}

	return info, nil
}

// Migrate database versions
func Migrate(queryer sql.Queryer, currentVersion string, versions []*Version) error {
	// Get migration info from the database
	info, err := GetMigrationInfo(queryer)
	if err != nil {
		return err
	}

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
		_, err = queryer.Exec("UPDATE _migration SET version=? WHERE id=?", currentVersion, info.ID)
		if err != nil {
			return err
		}
	} else if currentVersion < info.Version {
		// The current code version is lacking behind the database version, this is not allowed
		return errors.New("Database migration version is greater than current version")
	}

	return nil
}
