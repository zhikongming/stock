package utils

import (
	"fmt"
	"testing"
)

func TestGetPreShareholderReportDate(t *testing.T) {
	y := 2026
	d := 1
	for m := 1; m <= 12; m++ {
		dStr := fmt.Sprintf("%d", d)
		if d < 10 {
			dStr = fmt.Sprintf("0%s", dStr)
		}
		mStr := fmt.Sprintf("%d", m)
		if m < 10 {
			mStr = fmt.Sprintf("0%s", mStr)
		}
		dateStr := fmt.Sprintf("%d-%s-%s", y, mStr, dStr)
		date := ParseDate(dateStr)
		reportDate := GetShareholderReportDate(date)
		preReportDate := GetPreShareholderReportDate(reportDate)
		fmt.Printf("date: %s, reportDate: %s, preReportDate: %s\n", dateStr, reportDate, preReportDate)
	}
}

func TestGetShareholderNumberUnit(t *testing.T) {
	shareholderNumber := "12.50亿"
	number, unit := GetShareholderNumberUnit(shareholderNumber)
	fmt.Printf("shareholderNumber: %s, number: %f, unit: %s\n", shareholderNumber, number, unit)
	shareholderNumber2 := "3829万"
	number2, unit2 := GetShareholderNumberUnit(shareholderNumber2)
	fmt.Printf("shareholderNumber: %s, number: %f, unit: %s\n", shareholderNumber2, number2, unit2)
	shareholderNumber3 := "-3829万"
	number3, unit3 := GetShareholderNumberUnit(shareholderNumber3)
	fmt.Printf("shareholderNumber: %s, number: %f, unit: %s\n", shareholderNumber3, number3, unit3)
}
