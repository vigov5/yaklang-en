package suspect

import (
	"fmt"
	"strings"
)

// IsFullURL Guess whether it is a complete url based on value. Currently, we only care about http and https.
func IsFullURL(v interface{}) bool {
	var value = fmt.Sprint(v)
	prefix := []string{"http://", "https://"}
	value = strings.ToLower(value)
	for _, p := range prefix {
		if strings.HasPrefix(value, p) && len(value) > len(p) {
			return true
		}
	}
	return false
}

// Guess whether it is a url path based on value.
func IsURLPath(v interface{}) bool {
	var value = fmt.Sprint(v)
	return strings.Contains(value, "/") || commonURLPathExtRegex.MatchString(value)
}
