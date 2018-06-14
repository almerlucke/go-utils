package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	// DateTimeFormat SQL UTC datetime format used for all datetime format communication
	DateTimeFormat = "2006-01-02 15:04:05"
)

// DateTime time type alias for SQL datetime
type DateTime time.Time

// NewDateTime returns current UTC datetime
func NewDateTime() DateTime {
	return DateTime(time.Now().UTC())
}

// String stringer
func (t DateTime) String() string {
	return fmt.Sprintf("\"%v\"", time.Time(t).Format(DateTimeFormat))
}

/*
   Valuer interface for SQL driver
*/

// Value returns time.Time
func (t DateTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}

/*
   Scanner interface for SQL driver
*/

func (t *DateTime) scanString(s string) error {
	tt, err := time.Parse(DateTimeFormat, s)
	if err != nil {
		return err
	}

	*t = DateTime(tt.UTC())

	return nil
}

// Scan can scan []byte, string and time.Time
func (t *DateTime) Scan(src interface{}) error {
	// If value in db is NULL return current time
	if src == nil {
		*t = NewDateTime()
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
		*t = DateTime((src.(time.Time)).UTC())
	default:
		return errors.New("invalid src for sql.DateTime")
	}

	return nil
}

/*
	JSON marshal and unmarshal for sql.Time
*/

// MarshalJSON marshal sql.Time to json string
func (t DateTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%v\"", time.Time(t).Format(DateTimeFormat))), nil
}

// UnmarshalJSON unmarshal sql.Time from json string
func (t *DateTime) UnmarshalJSON(b []byte) error {
	var s string

	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	tt, err := time.Parse(DateTimeFormat, s)
	if err != nil {
		return err
	}

	*t = DateTime(tt.UTC())

	return nil
}
