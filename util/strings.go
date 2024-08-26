package util

// Contains searches slices for a string
func Contains(s []string, t string) bool {
	for _, e := range s {
		if e == t {
			return true
		}
	}
	return false
}
