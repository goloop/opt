package opt

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

// The sts converts slice to string.
// The function isn't intended for production and is used in testing
// when superficially comparing slices of different types.
//
// Examples:
//    sts([]int{1,2,3}, ",") == sts([]string{"1", "2", "3"}, ",") // true
func sts(slice interface{}, sep string) (r string) {
	switch reflect.TypeOf(slice).Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		s := reflect.ValueOf(slice)
		for i := 0; i < s.Len(); i++ {
			r += fmt.Sprint(s.Index(i)) + sep
		}
	}

	return strings.TrimSuffix(r, sep)
}

// The arg splits value by ':' and returns data as []string.
// The function isn't intended for production and is used in testing.
func arg(value string) []string {
	return strings.Split(value, ":")
}

// The split splits the string at the specified rune-marker ignoring the
// position of the marker inside of the group: `...`, '...', "..."
// and (...), {...}, [...].
//
// Examples:
//    split("a,b,c,d", ',')     // ["a", "b", "c", "d"]
//    split("a,(b,c),d", ',')   // ["a", "(b,c)", "d"]
//    split("'a,b',c,d", ',')   // ["'a,b'", "c", "d"]
//    split("a,\"b,c\",d", ',') // ["a", "\"b,c\"", "d"]
func split(str string, sep string) (r []string) {
	var (
		level int
		host  rune
		char  rune
		tmp   string

		flips    = map[rune]rune{'}': '{', ']': '[', ')': '('}
		quotes   = "\"'`"
		brackets = "({["
	)

	// The contains returns true if all items from the separators
	// were found in the string and it's all the same.
	contains := func(str string, separators ...rune) bool {
		var last = -1
		for _, sep := range separators {
			ir := strings.IndexRune(str, sep)
			if ir < 0 || (last >= 0 && last != ir) {
				return false
			}
			last = ir
		}

		return true
	}

	// Allocate the max memory size for storage all fields.
	r = make([]string, 0, strings.Count(str, ",")+1)

	// Split value.
	// for _, char = range str {
	for i := 0; i < utf8.RuneCountInString(str); i++ {
		char = rune(str[i])
		switch {
		case level == 0 && contains(quotes+brackets, char):
			host, level = char, level+1
		case contains(quotes, host, char):
			level, host = 0, 0
		case contains(brackets, host, flips[char]):
			level--
			if level <= 0 {
				level, host = 0, 0
			}
		case sep == str[i:i+utf8.RuneCountInString(sep)] && level == 0:
			i += utf8.RuneCountInString(sep) - 1
			r = append(r, tmp)
			tmp = ""
			continue
		}

		tmp += string(char)
	}

	// Add last piece to the result.
	if len(tmp) != 0 || string(char) == sep {
		r = append(r, tmp)
	}

	return
}
