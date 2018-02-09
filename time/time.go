// Package time contains time convenience methods and defines Unix timestamp.
package time

import (
	"math"
	"time"
)

// UnixTimestamp typedef for Unix timestamp in milliseconds
type UnixTimestamp int64

// StartOfDay truncate time to start of the day
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay ceil time to end of the day
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// Timestamp time to unix timestamp in milliseconds
func Timestamp(t time.Time) UnixTimestamp {
	return UnixTimestamp(t.UnixNano() / int64(time.Millisecond))
}

// Time convert timestamp to time
func (timestamp UnixTimestamp) Time() time.Time {
	seconds := float64(timestamp) / 1000.0
	nano := int64((seconds - math.Floor(seconds)) * float64(time.Second))
	return time.Unix(int64(seconds), nano)
}

// StartOfDay truncate timestamp to start of day
func (timestamp UnixTimestamp) StartOfDay() UnixTimestamp {
	return Timestamp(StartOfDay(timestamp.Time()))
}

// EndOfDay truncate timestamp to end of day
func (timestamp UnixTimestamp) EndOfDay() UnixTimestamp {
	return Timestamp(EndOfDay(timestamp.Time()))
}
