package helpers

import "fmt"

// Generic function to check if a parameter is nil
func IsNil[T comparable](value T) bool {
	// Use a trick to detect nil: compare to the zero value of T
	var zeroValue T
	return value == zeroValue && fmt.Sprintf("%v", value) == "<nil>"
}
