package utils

import "time"

const (
	DateFormat = "2006-01-02"
	TimeFormat = "2006-01-02 15:04:05"
)

func TimestampToDate(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format(DateFormat)
}

func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

func ParseDate(date string) time.Time {
	t, _ := time.Parse(DateFormat, date)
	return t
}

func ParseTime(timestamp string) time.Time {
	t, _ := time.ParseInLocation(TimeFormat, timestamp, time.Local)
	return t
}

func IsDateGreaterThan(date1, date2 string) bool {
	t1, _ := time.Parse(DateFormat, date1)
	t2, _ := time.Parse(DateFormat, date2)
	return t1.After(t2)
}
