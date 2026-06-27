package opt

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// The flagMap a special type that stores the names of flags and
// their number declared in the data structure.
type flagMap map[string]int

// The argValue is a special type to save the value of the argument.
type argValue struct {
	order int    // order of the argument on the command line
	value string // value for the argument
}

// looksLikeNumber reports whether s is a negative-number token such as
// "-5" or "-3.14". Such tokens must be treated as a value (or positional
// argument), not as a flag, so that options can receive negative numbers.
func looksLikeNumber(s string) bool {
	if len(s) < 2 || s[0] != '-' {
		return false
	}

	return (s[1] >= '0' && s[1] <= '9') || s[1] == '.'
}

// takeNextValue reports whether the next command-line token should be
// consumed as the value of the current option.
//
// A boolean option is a pure switch (Linux getopt style): it never consumes
// the following token, so "-v 5" keeps "5" as a positional argument. To set a
// boolean explicitly use "--no-flag", "--flag=false" or "-f=false". For any
// other option a token that looks like a flag is not a value, except for
// negative numbers (-5, -3.14).
func takeNextValue(next string, isBool bool) bool {
	if isBool {
		return false
	}

	return !strings.HasPrefix(next, "-") || looksLikeNumber(next)
}

// The argMap a special type of saving arguments in the form of a map.
//
// Save the list of values in the map not as []string but as []argValue
// so that you can reproduce the order of the elements in the list.
// For example, command line as: --users=John -UBob -U Roy, where -U is
// synonym for --users, can be passed as a slice [Bob Roy John] - i.e.
// not maintaining the order that was in the command line.
type argMap map[string][]argValue

// The parse converts the args slice to an argMap type.
//
// The shortFlags is a map of short flags which is used for
// separation value from the flag. Problem, for example: -dUGoloop,
// here -d and -U are short flags and Goloop is value for -U flag.
// How to understand that G is not another flag but the beginning
// of the value? This slice stores all available short flags
// declared in the data structure.
//
// The longFlags is a map of long flags. Problem, for example,
// the flag --no-verbose - is it the no-verbose flag or the objection
// for the verbose flag? This slice stores all available long flags
// declared in the data structure.
func (am argMap) parse(
	args []string,
	flags map[string]int,
	boolFlags map[string]bool,
) error {
	// Controls of parsing state of positional arguments.
	posState := struct {
		order  int  // real index of the positional argument
		active bool // true if switch to parsing of positional arguments
	}{}

	// Remove all keys from the map that could be there
	// from the previous parsing. We can't use am = make(argMap) here.
	clear(am)

	// Parse all the input arguments. Read elements on an index
	// instead of by means of range as sometimes it is necessary
	// to take the following argument in the course of current
	// iteration and to skip next one iteration.
	for i := 0; i < len(args); i++ {
		item := args[i]

		switch {
		case item == "--":
			// Toggles the parser to read positional arguments.
			//
			// Example:
			//   ./app 5 -UGoloop --no-verbose -d -- 10 15
			//
			// where 5, 10, 15 is positional arguments
			// and 10 isn't a value for -d flag.
			posState.active = true
		case strings.HasPrefix(item, "--") && !posState.active:
			// Long flag. The flag for a boolean value may have the no- prefix.
			// The value can be specified after the = symbol or as a series of
			// subsequent entries in the list. Values may also be missing for
			// boolean values.
			//
			// Example:
			//   ./app --user Goloop --no-verbose
			//   ./app --user=Goloop --verbose=false
			//   ./app --user Goloop
			//
			// where --user == "Goloop" and --verbose == false.
			var flag, data string
			var hasEq bool

			// Separate the flags from the value, like:
			// Example: --user=Goloop or --user Goloop where is user is a flag
			// and Goloop is a value. Or --no-verbose where verbose is a flag
			// and no- is boolean "no".

			// Split data and flag name.
			// Write the name of the long flag in lower case,
			// but we don't change the case for value.
			flag = strings.TrimPrefix(item, "--")
			if tmp := strings.SplitN(flag, "=", 2); len(tmp) == 2 {
				flag, data, hasEq = strings.ToLower(tmp[0]), tmp[1], true
			} else {
				flag = strings.ToLower(flag)
			}

			// Detect reverse mode and flag availability.
			switch _, ok := flags[flag]; {
			case !ok && strings.HasPrefix(flag, "no-"):
				// The flag is not available, but there is
				// a possibility that it needs reverse mode.
				fwn := strings.TrimPrefix(flag, "no-")
				if _, ok := flags[fwn]; ok {
					// The flag needs reverse mode but cannot be
					// reverse mode at the same time as data exists.
					if data == "" {
						flag, data = fwn, "false"
						break
					}
				}
				fallthrough
			case !ok:
				return fmt.Errorf("invalid argument %s", item)
			}

			key, value := flag, "true"
			if data != "" {
				value = string(data)
			} else if hasEq {
				// Explicit empty value: "--flag=" means an empty/zero
				// value, not a "true" switch and not the next token.
				value = ""
			} else if i+1 < len(args) && takeNextValue(
				args[i+1], boolFlags[key],
			) {
				// Take the value from the next item.
				value = args[i+1]
				i++ // be sure to move to the right by one position
			}

			am[key] = append(am[key], argValue{i, value})
		case strings.HasPrefix(item, "-") && !looksLikeNumber(item) &&
			!posState.active:
			// Short flag. The short flags can be grouped. The value can be
			// concatenated to the flag or as a series of subsequent entries
			// in the list.
			//
			// Example:
			//   ./app -dUGoloop -g"Hello, world"
			//   ./app -d -UGoloop -g "Hello, world"
			//   ./app -U Goloop -dg "Hello, world"
			//
			// where -d == true, -U == "Goloop" and -g == "Hello, world"
			var group, data []rune

			// Separate the flags from the value, like:
			// Example: -dUGoloop or -dU Goloop where is dU is the group of
			// flags and Goloop is value for -U, it is possible when d and U
			// in shortFlags and G isn't.
			//
			// In one group, a repeating flag is considered the result
			// of the previous flag. For example: -dUd, this is an alternative
			// to -dU d or -d -U d where last d is value for -U. Therefore, it
			// is necessary to monitor the uniqueness of the flag in the group
			// during parsing.
			unique := make(map[rune]bool)
			group = []rune(strings.TrimPrefix(item, "-"))
			for i, c := range group {
				_, exists := unique[c]
				if _, ok := flags[string(c)]; !ok || exists {
					group, data = group[:i], group[i:]
					break
				}
				unique[c] = true
			}

			// If no flag is set, this is a command line error.
			if len(group) == 0 {
				return fmt.Errorf("invalid argument %s", item)
			}

			for j, flag := range group {
				key, value := string(flag), "true"

				if j == len(group)-1 {
					// For last flag in flag list only.
					if len(data) != 0 {
						// Strip a leading "=" so that "-p=80" works
						// symmetrically with the long form "--port=80".
						value = strings.TrimLeft(string(data), " ")
						value = strings.TrimPrefix(value, "=")
					} else if i+1 < len(args) && takeNextValue(
						args[i+1], boolFlags[key],
					) {
						// Take the value from the next item.
						value = args[i+1]
						i++ // be sure to move to the right by one position
					}
				}

				am[key] = append(am[key], argValue{i, value})
			}
		default:
			// Value for the previous flag or positional arguments.
			// Positional arguments can be sets: before to the first
			// short or long flag; after the value to the last flag;
			// after switch as -- symbols.
			//
			// Example:
			//   ./app 5 10 15 -dUGoloop --verbose
			//   ./app 5 10 15 --verbose -dUGoloop
			//   ./app  --verbose -dUGoloop 5 10 15
			//   ./app  -dUGoloop --verbose -- 5 10 15
			//   ./app  5 10 -dUGoloop --verbose -- 15
			//
			// where 5, 10, 15 is positional arguments.
			am[strconv.Itoa(posState.order)] = []argValue{{i, item}}
			posState.order++
		}
	}

	return nil
}

// The asFlat returns argMap as simple map[string][]string.
func (am argMap) asFlat() map[string][]string {
	result := make(map[string][]string, len(am))

	for key, items := range am {
		tmp := make([]string, 0, len(items))
		for _, item := range items {
			tmp = append(tmp, item.value)
		}
		result[key] = tmp
	}

	return result
}

// The posValues returns values of positional arguments
// sorted in the sequence specified in the command line.
func (am argMap) posValues() []string {
	// Position argument object for sorting. A positional argument always
	// has exactly one value, e.g. "./app one two" -> "1": ["one"], etc.
	type posItem struct {
		order int    // real index of the positional argument
		value string // value of the positional argument
	}

	// Filter positional arguments only
	// (such arguments have a key that can be converted to an integer).
	tmp := make([]posItem, 0, len(am))
	for key, items := range am {
		order, err := strconv.Atoi(key)
		if err != nil || order == 0 || len(items) == 0 {
			continue // not a positional, or the app path ("0")
		}
		tmp = append(tmp, posItem{order, items[0].value})
	}

	// Arrange the items in the order specified in the command line.
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].order < tmp[j].order
	})

	// Return only the values of these arguments.
	result := make([]string, len(tmp))
	for i, item := range tmp {
		result[i] = item.value
	}

	return result
}

// The flagValue returns the value for the specified flag by
// long and/or short name with true as second param.
// Returns defValue with false as second param if the value
// is not found.
func (am argMap) flagValue(
	shortFlag,
	longFlag,
	defValue,
	sepList string,
) ([]string, bool) {
	short, long := am[shortFlag], am[longFlag]
	total := len(short) + len(long)

	switch total {
	case 0:
		// Return default value.
		return []string{defValue}, false
	case 1:
		// Single value from one of the synonyms: no merge or sort needed
		// (the common case for non-list fields).
		if len(short) == 1 {
			return []string{short[0].value}, true
		}
		return []string{long[0].value}, true
	}

	// Merge values from both synonyms and restore the sequence that was on
	// the command line. For example: -UOne --users=Two -U Three where U and
	// users are synonyms, the result must be ["One", "Two", "Three"] and not
	// ["One", "Three", "Two"] (short flags are stored first).
	tmp := make([]argValue, 0, total)
	tmp = append(tmp, short...)
	tmp = append(tmp, long...)
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].order < tmp[j].order
	})

	result := make([]string, len(tmp))
	for i, item := range tmp {
		result[i] = item.value
	}

	return result, true
}
