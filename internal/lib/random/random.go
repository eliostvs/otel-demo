package random

import (
	"math/rand"
)

// Normalvariate is the normal distribution.
func Normalvariate(mean, sigma float64) float64 {
	return rand.NormFloat64()*sigma + mean
}

// Choice returns a random value from a slice.
func Choice[T any](array []T) T {
	idx := rand.Intn(len(array))
	return array[idx]
}

func IntRange(min, max int) int {
	return rand.Intn(max-min) + min
}
