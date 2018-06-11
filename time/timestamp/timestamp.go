// Package timestamp contains time convenience methods and defines Unix timestamp.
package timestamp

import (
	"math"
	"time"

	timeUtils "github.com/almerlucke/go-utils/time"
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

// StartOfDay truncate timestamp to start of day
func (timestamp Timestamp) StartOfDay() Timestamp {
	return New(timeUtils.StartOfDay(timestamp.Time()))
}

// EndOfDay truncate timestamp to end of day
func (timestamp Timestamp) EndOfDay() Timestamp {
	return New(timeUtils.EndOfDay(timestamp.Time()))
}
