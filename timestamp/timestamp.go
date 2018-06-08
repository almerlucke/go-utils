// Package timestamp contains time convenience methods and defines Unix timestamp.
package timestamp

import (
	"math"
	"time"
)

// Timestamp typedef for Unix timestamp in milliseconds
type Timestamp int64

// New time to unix timestamp in milliseconds
func New(t time.Time) Timestamp {
	return Timestamp(t.UnixNano() / int64(time.Millisecond))
}

// Time convert timestamp to time
func (timestamp Timestamp) Time() time.Time {
	seconds := float64(timestamp) / 1000.0
	nano := int64((seconds - math.Floor(seconds)) * float64(time.Second))
	return time.Unix(int64(seconds), nano)
}

// StartOfDay truncate time to start of the day
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay ceil time to end of the day
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfDay truncate timestamp to start of day
func (timestamp Timestamp) StartOfDay() Timestamp {
	return New(StartOfDay(timestamp.Time()))
}

// EndOfDay truncate timestamp to end of day
func (timestamp Timestamp) EndOfDay() Timestamp {
	return New(EndOfDay(timestamp.Time()))
}
