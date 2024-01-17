package utils

import (
	"regexp"

	"github.com/gobwas/glob"
)

func interfaceToStr(i interface{}) string {
	return InterfaceToString(i)
}

// MatchAnyOfSubString tries to convert i to a string, and then determines whether any substring subStr exists in i. If one of the substrings exists in i, it returns true, otherwise it returns false. This function ignores the case.
// Example:
// ```
// str.MatchAnyOfSubString("abc", "a", "z", "x") // true
// ```
func MatchAnyOfSubString(i interface{}, subStr ...string) bool {
	raw := interfaceToStr(i)
	for _, subStr := range subStr {
		if IContains(raw, subStr) {
			return true
		}
	}
	return false
}

// MatchAllOfSubString tries to convert i Convert to a string, and then determine whether all substrings subStr exist in i, return true if they exist, otherwise return false, this function ignores case
// Example:
// ```
// str.MatchAllOfSubString("abc", "a", "b", "c") // true
// ```
func MatchAllOfSubString(i interface{}, subStr ...string) bool {
	if len(subStr) <= 0 {
		return false
	}

	raw := interfaceToStr(i)
	for _, subStr := range subStr {
		if !IContains(raw, subStr) {
			return false
		}
	}
	return true
}

// MatchAnyOfGlob Try to convert i to a string, and then use glob matching pattern matching , if any glob pattern matches successfully, return true, otherwise return false
// Example:
// ```
// str.MatchAnyOfGlob("abc", "a*", "??b", "[^a-z]?c") // true
// ```
func MatchAnyOfGlob(
	i interface{}, re ...string) bool {
	raw := interfaceToStr(i)
	for _, r := range re {
		if glob.MustCompile(r).Match(raw) {
			return true
		}
	}
	return false
}

// MatchAllOfGlob Try to convert i to string, then use glob matching pattern to match, if all glob patterns If all match successfully, it returns true, otherwise it returns false.
// Example:
// ```
// str.MatchAllOfGlob("abc", "a*", "?b?", "[a-z]?c") // true
// ```
func MatchAllOfGlob(
	i interface{}, re ...string) bool {
	if len(re) <= 0 {
		return false
	}

	raw := interfaceToStr(i)
	for _, r := range re {
		if !glob.MustCompile(r).Match(raw) {
			return false
		}
	}
	return true
}

// MatchAnyOfRegexp tries to convert i Convert to string, then use regular expression matching, return true if any regular expression matches successfully, otherwise return false
// Example:
// ```
// str.MatchAnyOfRegexp("abc", "a.+", "Ab.?", ".?bC") // true
// ```
func MatchAnyOfRegexp(
	i interface{},
	re ...string) bool {
	raw := interfaceToStr(i)
	for _, r := range re {
		result, err := regexp.MatchString(r, raw)
		if err != nil {
			continue
		}
		if result {
			return true
		}
	}
	return false
}

// MatchAllOfRegexp tries to convert i into a string, and then uses regular expression matching. If all regular expressions match successfully, it returns true, otherwise it returns false.
// Example:
// ```
// str.MatchAllOfRegexp("abc", "a.+", ".?b.?", "\\w{2}c") // true
// ```
func MatchAllOfRegexp(
	i interface{},
	re ...string) bool {
	if len(re) <= 0 {
		return false
	}

	raw := interfaceToStr(i)
	for _, r := range re {
		result, err := regexp.MatchString(r, raw)
		if err != nil {
			return false
		}
		if !result {
			return false
		}
	}
	return true
}
