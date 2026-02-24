package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

func IsStockNumber(code string) bool {
	re := regexp.MustCompile(`^\d+$`)
	return len(code) == 6 && re.MatchString(code)
}

func IsStockCodeWithPrefix(code string) bool {
	if len(code) != 8 {
		return false
	}
	prefix := strings.ToUpper(code[:2])
	_, ok := StockIdMap[prefix]
	return ok
}

func IsIndustryCode(code string) bool {
	return len(code) == 6 && strings.HasPrefix(code, "BK")
}

func GetFullStockCodeOfNumber(code string) string {
	for matchPrefix, codePrefix := range CodeToPrefixMap {
		if strings.HasPrefix(code, matchPrefix) {
			return fmt.Sprintf("%s%s", codePrefix, code)
		}
	}
	return code
}

func GetStockCodeNumber(code string) string {
	if IsStockNumber(code) {
		return code
	}
	if len(code) == 8 {
		return code[2:]
	}
	return code
}

func RemoveIndustryNumberSuffix(name string) string {
	return strings.TrimRightFunc(name, unicode.IsNumber)
}
