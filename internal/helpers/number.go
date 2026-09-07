package helpers

import "math"

func Rounded(x float64, d int) float64 {
	factor := math.Pow(10, float64(d))
	return math.Round(x*factor) / factor
}

func RemoveDuplicateInt(intSlice []int) []int {
	allKeys := make(map[int]bool)
	list := []int{}
	for _, item := range intSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
