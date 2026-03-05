package utils

import (
	"testing"
)

func TestIsWeekend(t *testing.T) {
	t.Run("", func(t *testing.T) {
		if got := IsNowWeekend(); got != true {
			t.Errorf("IsNowWeekend() = %v, want %v", got, true)
		}
	})
}
