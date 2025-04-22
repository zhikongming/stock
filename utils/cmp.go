package utils

func IsClosedToHigh(data, high, low, percent float64) bool {
	if data >= high {
		return true
	} else if data <= low {
		return false
	} else {
		return data >= (high-low)*percent+low
	}
}
