package helpers

func ArrayFindIndex[T any](slice []T, predicate func(T) bool) int {
	for i, v := range slice {
		if predicate(v) {
			return i
		}
	}
	return -1 // Return -1 if no match is found
}

func ArrayFilter[T any](arr []T, condition func(T) bool) []T {
	var result []T
	for _, v := range arr {
		if condition(v) {
			result = append(result, v)
		}
	}
	return result
}
