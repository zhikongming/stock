package utils

func ListSwap[T any](list []T) []T {
	length := len(list)
	if length <= 1 {
		return list
	}
	for i := 0; i < length/2; i++ {
		list[i], list[length-i-1] = list[length-i-1], list[i]
	}
	return list
}

func ListFloat64Average(list []float64) float64 {
	if len(list) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range list {
		sum += v
	}
	return sum / float64(len(list))
}
