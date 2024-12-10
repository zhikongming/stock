package utils

import "time"

const (
	DateFormat = "2006-01-02"
)

func TimestampToDate(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format(DateFormat)
}
