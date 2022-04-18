package random

import (
	"math/rand"

	"github.com/username/otel-playground/internal/lib/constraints"
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

func NumberInRange[T constraints.Numbers](min, max T) T {
	return T(rand.Intn(int(max)-int(min)) + int(min))
}
