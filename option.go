package opt

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Option identification constants.
const (
	defaultOptionValue = "true"   // default value for flags
	notOption          = iota + 0 // not an option
	shortOption                   // short option starts from one `-`
	longOption                    // long option starts form two `--`
	switchOption                  // switch for entering positional options
)

// Rules for handling options.
var (
	positionalKeyRgx = regexp.MustCompile(`^\d+$`)      // switch option
	flagOptionRgx    = regexp.MustCompile(`^(\s*)-\w`)  // short option
	cmdOptionRgx     = regexp.MustCompile(`^(\s*)--\w`) // long option
	switchOptionRgx  = regexp.MustCompile(`^(\s*)--$`)  // switch option
)

// The optSamples is map as opt:value.
type optSamples map[string]string

// PositionalValues returns positional argument values as list.
func (oss optSamples) PositionalValues() []string {
	var r = []string{}

	// Order of the arguments must be respected.
	// Find positional keys.
	for key := range oss {
		// Ignore 0 position - path to bin file.
		if key != "0" && positionalKeyRgx.Match([]byte(key)) {
			r = append(r, key)
		}
	}

	// Sort positional keys and revrite values.
	sort.Strings(r)
	for i, key := range r {
		r[i] = oss[key]
	}

	return r
}

// The isSep returns true if runce is separator. The sops is a string of
// characters (short option names as string) that cannot be as separators.
func isSep(v rune, sops string) bool {
	if len(sops) == 0 {
		// The not-separator list is't defined. So, the separator
		// is everything that is not in the A-Z,a-z range.
		return ((v < 'a' || v > 'z') && (v < 'A' || v > 'Z'))
	} else if strings.Contains(sops, string(v)) {
		// The value in not-separator list.
		return false
	}

	return true
}

// The getOptionKind returns option identification constant for arg.
func getOptionKind(arg string) int {
	switch {
	case flagOptionRgx.Match([]byte(arg)):
		return shortOption
	case cmdOptionRgx.Match([]byte(arg)):
		return longOption
	case switchOptionRgx.Match([]byte(arg)):
		return switchOption
	}

	return notOption
}

// The parseOptions parses the argument list. Where args is argument list
// like os.Args and sops is string with names of the short options.
func getOptions(args []string, sops string) (r optSamples, err error) {
	// The position is structure for definition positional options.
	type position struct {
		id      int  // position index
		active  bool // if true - read positional options
		blocked bool // if true - cannot switch active
	}

	var pos = position{0, true, false} // current position argument

	// Check all arguments.
	r = optSamples{}
	for i := 0; i < len(args); i++ {
		var (
			arg  = args[i]
			kind = getOptionKind(arg)
		)

		switch {
		case kind == shortOption && !pos.blocked:
			var name, value string
			name = strings.TrimPrefix(arg, "-")
			pos.active = false

			// Define a group of options and the value of the last option.
			for n := 0; n < len(name); n++ {
				if isSep(rune(name[n]), sops) {
					name, value = name[:n], name[n:]
					break
				}
			}

			// The name not exists - ignore this arg.
			if len(name) == 0 {
				r = optSamples{}
				err = fmt.Errorf("option name is incorrect %s", arg)
				return
			}

			// Perhaps the following argument is a value.
			if len(value) == 0 {
				value = defaultOptionValue // set default value

				// Try to get the next opt as value.
				if len(args) > i+1 && getOptionKind(args[i+1]) == notOption {
					value = args[i+1]
					i++
				}
			}

			// Name can be as group of short options.
			r[string(name[len(name)-1])] = strings.Trim(value, " ")
			for n := 0; n < len(name)-1; n++ {
				r[string(name[n])] = defaultOptionValue
			}
		case kind == longOption && !pos.blocked:
			var name, value string
			name = strings.ToLower(strings.TrimPrefix(arg, "--"))
			pos.active = false

			// Minimum two chars for long option (without denial index `no-`).
			if len(strings.TrimPrefix(name, "no-")) < 2 {
				r = optSamples{}
				err = fmt.Errorf("option name is incorrect %s", arg)
				return
			}

			// Split to left and right.
			for n := 0; n < len(name); n++ {
				if strings.Contains("= ", string(name[n])) {
					items := strings.SplitN(name, string(name[n]), 2)
					if len(items) == 2 {
						name, value = items[0], items[1]
					}
					break
				}
			}

			// Perhaps the following argument is a value.
			if len(value) == 0 {
				value = defaultOptionValue // set default value

				// Try to get the next opt as value.
				if len(args) > i+1 && getOptionKind(args[i+1]) == notOption {
					value = args[i+1]
					i++
				}
			}

			// If name has "no-" prefix - set `false` for value.
			if strings.HasPrefix(name, "no-") {
				name = strings.TrimPrefix(name, "no-")
				value = "false"
			}

			// Save flag in lowercase.
			r[name] = strings.Trim(value, " ")
		case kind == switchOption:
			// Positional arguments sets after the options list and
			// separated by empty marker as `--`.
			// Note: Change state and ignore this step.
			pos.active = true
			pos.blocked = true
		default:
			if pos.id != 0 && !pos.active {
				// Positional arguments sets without
				// a special delimiter as `--`.
				// Note: Change state and continue parsing.
				pos.active = true
				pos.blocked = true
			}

			if pos.active {
				// Positional arguments written by index: 0, 1, 2, ..., .
				r[fmt.Sprintf("%d", pos.id)] = arg
				pos.id++
			} else {
				// Invalid option.
				r = optSamples{}
				err = fmt.Errorf("option name is incorrect %s", arg)
				return
			}
		} // case
	} // for

	return
}
