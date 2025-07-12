package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

func ToString(i interface{}) string {
	return fmt.Sprintf("%v", i)
}

func ToFloat64(v interface{}) float64 {
	if fv, ok := v.(float64); ok {
		return fv
	}

	switch v.(type) {
	case json.Number:
		if fv, err := v.(json.Number).Float64(); err == nil {
			return fv
		} else {
			return 0
		}
	case float32:
		return float64(v.(float32))
	case int, int8, int16, int32, int64:
		return float64(v.(int64))
	case uint, uint8, uint16, uint32, uint64:
		return float64(v.(uint64))
	case string:
		if fv, err := strconv.ParseFloat(v.(string), 64); err == nil {
			return fv
		} else {
			return 0
		}
	default:
		return 0
	}
}

func ToInt64(v interface{}) int64 {
	if iv, ok := v.(int64); ok {
		return iv
	}

	switch v.(type) {
	case json.Number:
		if iv, err := v.(json.Number).Int64(); err == nil {
			return iv
		} else {
			return 0
		}
	case int, int8, int16, int32:
		return int64(v.(int))
	case uint, uint8, uint16, uint32:
		ui := v.(uint64)
		if ui > uint64(int64(ui)) {
			return 0
		}
		return int64(ui)
	case string:
		if iv, err := strconv.ParseInt(v.(string), 10, 64); err == nil {
			return iv
		} else {
			return 0
		}
	case float64:
		fv := v.(float64)
		if fv < float64(int64(fv)) && fv > -float64(int64(fv)) {
			return int64(fv)
		} else {
			return 0
		}
	default:
		return 0
	}
}

func Float64KeepDecimal(f float64, decimal int) float64 {
	return math.Round(f*math.Pow10(decimal)) / math.Pow10(decimal)
}

func Float64Equal(f1, f2 float64, decimal int) bool {
	f1 = Float64KeepDecimal(f1, decimal)
	f2 = Float64KeepDecimal(f2, decimal)
	return f1 == f2
}
