package service

import (
	"fmt"
	"testing"
)

func TestIsLimitUpWithRate(t *testing.T) {
	code := "600481"
	rate := GetLimitUpRate(code)
	prevClosePrice := 6.33
	curClosePrice := 6.96
	fmt.Printf("IsLimitUpWithRate = %v\n", IsLimitUpWithRate(prevClosePrice, curClosePrice, rate))

	code = "600903"
	rate = GetLimitUpRate(code)
	prevClosePrice = 6.34
	curClosePrice = 6.97
	fmt.Printf("IsLimitUpWithRate = %v\n", IsLimitUpWithRate(prevClosePrice, curClosePrice, rate))
	prevClosePrice = 6.97
	curClosePrice = 7.67
	fmt.Printf("IsLimitUpWithRate = %v\n", IsLimitUpWithRate(prevClosePrice, curClosePrice, rate))

	code = "920580"
	rate = GetLimitUpRate(code)
	prevClosePrice = 14.49
	curClosePrice = 18.83
	fmt.Printf("IsLimitUpWithRate = %v\n", IsLimitUpWithRate(prevClosePrice, curClosePrice, rate))
}
