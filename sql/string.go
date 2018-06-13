package sql

import (
	"database/sql/driver"
	"errors"
)

// String for DB, set to "" if db field is NULL
type String string

// Value - Implementation of valuer for database/sql
func (s String) Value() (driver.Value, error) {
	// value needs to be a base driver.Value type
	// such as string.
	return string(s), nil
}

// Scan sql string, if NULL string is set to empty string
func (s *String) Scan(value interface{}) error {
	// if value is nil, false
	if value == nil {
		*s = String("")
		return nil
	}

	switch value.(type) {
	case string:
		*s = String(value.(string))
		return nil
	case []byte:
		*s = String(value.([]byte))
		return nil
	}

	// otherwise, return an error
	return errors.New("failed to scan sql.String")
}
