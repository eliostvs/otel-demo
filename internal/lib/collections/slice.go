package collections

func SliceContains[T comparable](x T, arr []T) bool {
	for _, v := range arr {
		if v == x {
			return true
		}
	}
	return false
}
