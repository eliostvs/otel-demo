package math

import (
	"github.com/username/otel-playground/internal/lib/constraints"
)

// Max returns the biggest of two numbers
func Max[T constraints.Numbers](x, y T) T {
	if x > y {
		return x
	}
	return y
}
