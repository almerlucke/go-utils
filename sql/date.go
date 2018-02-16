package sql

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	// DateFormat SQL UTC datetime format used for all date format communication
	DateFormat = "2006-01-02"
)

// Date time type alias for SQL date
type Date time.Time

// NewDate returns current UTC date
func NewDate() Date {
	return Date(time.Now().UTC())
}

// String stringer
func (t Date) String() string {
	return fmt.Sprintf("\"%v\"", time.Time(t).Format(DateFormat))
}

/*
   Valuer interface for SQL driver
*/

// Value returns time.Time
func (t Date) Value() (driver.Value, error) {
	return time.Time(t), nil
}

/*
   Scanner interface for SQL driver
*/

func (t *Date) scanString(s string) error {
	tt, err := time.Parse(DateFormat, s)
	if err != nil {
		return err
	}

	*t = Date(tt)

	return nil
}

// Scan can scan []byte, string and time.Time
func (t *Date) Scan(src interface{}) error {
	// If value in db is NULL return current time
	if src == nil {
		*t = Date(time.Now())
		return nil
	}

	switch src.(type) {
	case []byte:
		err := t.scanString(string(src.([]byte)))
		if err != nil {
			return err
		}
	case string:
		err := t.scanString(src.(string))
		if err != nil {
			return err
		}
	case time.Time:
		*t = Date(src.(time.Time))
	default:
		return errors.New("Invalid src for sql.Date")
	}

	return nil
}

/*
	JSON marshal and unmarshal for sql.Time
*/

// MarshalJSON marshal sql.Date to json string
func (t Date) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%v\"", time.Time(t).Format(DateFormat))), nil
}

// UnmarshalJSON unmarshal sql.Date from json string
func (t *Date) UnmarshalJSON(b []byte) error {
	var s string

	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	tt, err := time.Parse(DateFormat, s)
	if err != nil {
		return err
	}

	*t = Date(tt)

	return nil
}
