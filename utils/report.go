package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func GetPreShareholderReportDate(reportDate string) string {
	// 解析日期字符串
	date := ParseDate(reportDate)

	year := date.Year()
	month := date.Month()
	day := date.Day()

	// 根据月日判断上一份财报日期
	switch {
	case month == 9 && day == 30:
		// 09-30 的上一份财报是 06-30
		return fmt.Sprintf("%d-06-30", year)
	case month == 6 && day == 30:
		// 06-30 的上一份财报是 03-31
		return fmt.Sprintf("%d-03-31", year)
	case month == 3 && day == 31:
		// 03-31 的上一份财报是前一年的 12-31
		return fmt.Sprintf("%d-12-31", year-1)
	default:
		// 默认返回空字符串，表示无法确定上一份财报日期
		return ""
	}
}

func GetShareholderReportDate(date time.Time) string {
	year := date.Year()
	month := date.Month()

	switch {
	case month >= 1 && month <= 4:
		// 1-4月，返回去年的12月31日
		return fmt.Sprintf("%d-09-30", year-1)
	case month >= 5 && month <= 8:
		// 5-8月，返回今年的3月31日
		return fmt.Sprintf("%d-03-31", year)
	case month >= 9 && month <= 10:
		// 9-12月，返回今年的6月30日
		return fmt.Sprintf("%d-06-30", year)
	case month >= 11 && month <= 12:
		return fmt.Sprintf("%d-09-30", year)
	default:
		return fmt.Sprintf("%d-12-31", year-1)
	}
}

func GetShareholderNumberUnit(shareholderNumber string) (float64, string) {
	num := "0.0"
	unit := ""
	re := regexp.MustCompile(`^(-?\d+(?:\.\d+)?)(.*)$`)
	matches := re.FindStringSubmatch(shareholderNumber)
	if len(matches) == 3 {
		num, unit = matches[1], matches[2]
	}
	number, _ := strconv.ParseFloat(num, 64)
	return number, unit
}

func GetGetShareholderNumberByUnit(unit string) int {
	switch unit {
	case "万":
		return 10000
	case "亿":
		return 1000000000
	default:
		return 1
	}
}
