package opt

import "strings"

// The split splits the string on elements using a colon as a separator.
//
// Note: used exclusively for testing.
func split(str string) []string {
	return strings.Split(str, ":")
}
