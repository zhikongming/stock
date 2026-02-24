package utils

import (
	"testing"
)

func TestRemoveIndustryNumberSuffix(t *testing.T) {
	origin := "白酒Ⅲ"
	t.Run(origin, func(t *testing.T) {
		if got := RemoveIndustryNumberSuffix(origin); got != "白酒" {
			t.Errorf("RemoveIndustryNumberSuffix() = %v, want %v", got, "白酒")
		}
	})
}
