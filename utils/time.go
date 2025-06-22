package utils

import "time"

const (
	DateFormat      = "2006-01-02"
	TimeFormat      = "2006-01-02 15:04:05"
	Date2Format     = "20060102"
	TimeShortFormat = "2006-01-02 15:04"
)

func TimestampToDate(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format(DateFormat)
}

func TimestampToDateTime(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format(TimeFormat)
}

func TimeToTimestamp(t time.Time) int64 {
	return t.Unix()
}

func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

func FormatDate2(t time.Time) string {
	return t.Format(Date2Format)
}

func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

func FormatShortTime(t time.Time) string {
	return t.Format(TimeShortFormat)
}

func ParseDate(date string) time.Time {
	t, _ := time.Parse(DateFormat, date)
	return t
}

func ParseDate2(date string) time.Time {
	t, _ := time.Parse(Date2Format, date)
	return t
}

func ParseTime(timestamp string) time.Time {
	t, _ := time.ParseInLocation(TimeFormat, timestamp, time.Local)
	return t
}

func ParseShortTime(timestamp string) time.Time {
	t, _ := time.ParseInLocation(TimeShortFormat, timestamp, time.Local)
	return t
}

func IsDateGreaterThan(date1, date2 string) bool {
	t1, _ := time.Parse(DateFormat, date1)
	t2, _ := time.Parse(DateFormat, date2)
	return t1.After(t2)
}
