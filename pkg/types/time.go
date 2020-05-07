package types

import "time"

// Format used by db to store timestamps
const Format = "2006-01-02T15:04:05"

// this is needed in order for our API library to work with hasura
type Timestamp string

// helper function which always returns valid string and never panics
func ToTimestamp(t time.Time) Timestamp {
	return Timestamp(t.Format(Format))
}

// helper function which returns time object from timestamp and never panics
func FromTimestamp(s Timestamp) time.Time {
	t, _ := time.Parse(Format, string(s))
	return t
}
