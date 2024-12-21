package validate

func IsSubset(s string) bool {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	charSet := make(map[rune]struct{})
	for _, ch := range chars {
		charSet[ch] = struct{}{}
	}

	for _, ch := range s {
		if _, exists := charSet[ch]; !exists {
			return false
		}
	}
	return true
}
