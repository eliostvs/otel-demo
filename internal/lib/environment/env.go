package environment

import (
	"os"
	"regexp"
	"strconv"
)

// Get returns the given environment variable or the fallback value
func Get[T any](key string, fallback T) T {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	var ret any
	switch any(fallback).(type) {
	case string:
		ret = value

	case float64:
		f, _ := strconv.ParseFloat(value, 64)
		ret = f

	case int64:
		i, _ := strconv.ParseInt(value, 10, 64)
		ret = i

	case int:
		i, _ := strconv.ParseInt(value, 10, 64)
		ret = int(i)

	case bool:
		b, _ := regexp.Match(`(?i)(false|0)`, []byte(value))
		ret = b
	}

	return ret.(T)
}
