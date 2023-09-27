package common

func Includes[T comparable](elems []T, v T) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}

	return false
}

func ValueOrDefault(v *string, d string) string {
	if v == nil || *v == "" {
		return d
	}

	return *v
}
