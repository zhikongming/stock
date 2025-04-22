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
