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

func ListSum[T int64 | float64 | int32 | int](list []T) T {
	if len(list) == 0 {
		return 0
	}
	sum := T(0)
	for _, v := range list {
		sum += v
	}
	return sum
}

func In[T comparable](item T, list []T) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

func ListStringIgnoreEmpty(list []string) []string {
	var result []string
	for _, v := range list {
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}

func Uniq[T comparable](list []T) []T {
	localMap := make(map[T]struct{})
	for _, item := range list {
		localMap[item] = struct{}{}
	}
	uniqList := make([]T, 0, len(localMap))
	for item := range localMap {
		uniqList = append(uniqList, item)
	}
	return uniqList
}
