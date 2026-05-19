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
	fmt.Printf("%s: IsLimitUpWithRate = %v\n", code, IsLimitUpWithRate(prevClosePrice, curClosePrice, rate))

	code = "600903"
	rate = GetLimitUpRate(code)
	prevClosePrice = 6.34
	curClosePrice = 6.97
	fmt.Printf("%s: IsLimitUpWithRate = %v\n", code, IsLimitUpWithRate(prevClosePrice, curClosePrice, rate))
	prevClosePrice = 6.97
	curClosePrice = 7.67
	fmt.Printf("%s: IsLimitUpWithRate = %v\n", code, IsLimitUpWithRate(prevClosePrice, curClosePrice, rate))

	code = "920580"
	rate = GetLimitUpRate(code)
	prevClosePrice = 14.49
	curClosePrice = 18.83
	fmt.Printf("%s: IsLimitUpWithRate = %v\n", code, IsLimitUpWithRate(prevClosePrice, curClosePrice, rate))

	code = "603311"
	rate = GetLimitUpRate(code)
	prevClosePrice = 21.00
	curClosePrice = 23.10
	fmt.Printf("%s: IsLimitUpWithRate = %v\n", code, IsLimitUpWithRate(prevClosePrice, curClosePrice, rate))

	code = "603311"
	rate = GetLimitUpRate(code)
	prevClosePrice = 23.10
	curClosePrice = 25.41
	fmt.Printf("%s: IsLimitUpWithRate = %v\n", code, IsLimitUpWithRate(prevClosePrice, curClosePrice, rate))

	code = "603311"
	rate = GetLimitUpRate(code)
	prevClosePrice = 25.41
	curClosePrice = 27.95
	fmt.Printf("%s: IsLimitUpWithRate = %v\n", code, IsLimitUpWithRate(prevClosePrice, curClosePrice, rate))
}
