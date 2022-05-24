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
// func (am argMap) parse(args []string, shortFlags, longFlags flagMap) error {
func (am argMap) parse(args []string, flags map[string]int) error {
	// Controls of parsing state of positional arguments.
	var posState = struct {
		order  int  // real index of the positional argument
		active bool // true if switch to parsing of positional arguments
	}{}

	// Remove all keys from the map that could be there
	// from the previous parsing. We can't use am = make(argMap) here.
	for key := range am {
		delete(am, key)
	}

	// Parse all the input arguments. Read elements on an index
	// instead of by means of range as sometimes it is necessary
	// to take the following argument in the course of current
	// iteration and to skip next one iteration.
	for i := 0; i < len(args); i++ {
		var item = args[i]

		switch {
		case item == "--":
			// Toggles the parser to read positional arguments.
			//
			// Example:
			//   ./app 5 -UGoloop --verbose false -d -- 10 15
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
			//   ./app --user Goloop --verbose false
			//   ./app --user Goloop
			//
			// where --user == "Goloop" and --verbose == false.
			var flag, data string

			// Separate the flags from the value, like:
			// Example: --user=Goloop or --user Goloop where is user is a flag
			// and Goloop is a value. Or --no-verbose where verbose is a flag
			// and no- is boolean "no".

			// Split data and flag name.
			// Write the name of the long flag in lower case,
			// but we don't change the case for value.
			flag = strings.TrimPrefix(item, "--")
			if tmp := strings.SplitN(flag, "=", 2); len(tmp) == 2 {
				flag, data = strings.ToLower(tmp[0]), tmp[1]
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
			} else if i+1 < len(args) {
				// Try to take the value from the next item.
				if tmp := args[i+1]; !strings.HasPrefix(tmp, "-") {
					value = tmp
					i++ // be sure to move to the right by one position
				}
			}

			am[key] = append(am[key], argValue{i, value})
		case strings.HasPrefix(item, "-") && !posState.active:
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
						value = strings.TrimLeft(string(data), " ")
					} else if i+1 < len(args) {
						// Try to take the value from the next item.
						if tmp := args[i+1]; !strings.HasPrefix(tmp, "-") {
							value = tmp
							i++ // be sure to move to the right by one position
						}
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
			am[fmt.Sprint(posState.order)] = []argValue{{i, item}}
			posState.order++
		}
	}

	return nil
}

// The asFlat returns argMap as simple map[string][]string.
func (am argMap) asFlat() map[string][]string {
	var result = make(map[string][]string, len(am))

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
	// Position argument object for sorting.
	type posItem struct {
		order int      // real index of the positional argument
		value []string // value of the posipositionaltion argument
	}

	// Filter positional arguments only
	// (such arguments have a key that can be converted to an integer).
	tmp := []posItem{}
	for key, items := range am {
		if order, err := strconv.Atoi(key); err == nil {
			if order == 0 {
				continue // name of app
			}
			value := make([]string, 0, len(items))
			for _, item := range items {
				value = append(value, item.value)
			}
			tmp = append(tmp, posItem{order, value})
		}
	}

	// Arrange the items in the order specified in the command line.
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].order < tmp[j].order
	})

	// Return only the values of these arguments.
	result := make([]string, 0, len(tmp))
	for _, item := range tmp {
		// A positional argument can have only one value, for example:
		// ./app one two three "Hello World" five  - five positional arguments,
		// but each has one meaning without exception:
		// "0": []string{"./app"},
		// "1": []string{"one"},
		// "2": []string{"two"},
		// ...,
		// "5": []string{"five"}.
		//
		// if len(item.value) != 1 {
		//     panic("this event will never happen")
		// }
		result = append(result, item.value[0])
	}

	return result
}

// The flagValue returns the value for the specified flag by
// long and/or short name. If the value is not found - returns defValue.
func (am argMap) flagValue(
	shortFlag string,
	longFlag string,
	defValue string,
	sepList string,
) []string {
	var result []string

	// Join all values from all types of flags (long/short).
	tmp := []argValue{}
	for _, key := range []string{shortFlag, longFlag} {
		if items, ok := am[key]; ok {
			tmp = append(tmp, items...)
		}
	}

	// Return default value.
	if len(tmp) == 0 {
		return []string{defValue}
	}

	// Be sure to set the sequence of values that was in the command line.
	// For example: -UOne --users=Two -U Three where U and users is synonyms,
	// result should be ["One", "Two", "Three"], but now it will
	// look like ["One", "Three", "Two"] because short flags
	// are processed first.
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].order < tmp[j].order
	})

	// Get the value of the arguments only.
	for _, item := range tmp {
		result = append(result, item.value)
	}

	return result
}
