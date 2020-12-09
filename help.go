package opt

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

// Helper is the interface implemented by types that can
// returns help information about app.
type Helper interface {
	HelpOPT(string) string
}

// Usager is the interface implemented by types that can
// returns information about using command line parameters.
type Usager interface {
	UsageOPT(string) string
}

// posArgsNum structure for calculating positional arguments and
// create a text specification about it.
type posArgsNum struct {
	exists bool // true if positional arguments is exists
	array  int  // array of positional arguments
	enum   int  // enumerated positional argument counter
}

// Count counts the number of positional arguments and
// returns true if a count has been made.
func (pan *posArgsNum) Count(field *fieldSample) bool {
	switch {
	case field.TagSample.Short == "?" || field.TagSample.Short == "0":
		// Ignore zero positional argument and help data field.
		return true
	case field.TagSample.Short == "[]":
		// There is a slice for positional arguments.
		if field.Item.Kind() == reflect.Array {
			// There is an array - to determine max positional
			// arguments number.
			pan.array = field.Item.Type().Len()
		}
		pan.exists = true
		return true
	case positionalKeyRgx.Match([]byte(field.TagSample.Short)):
		// Enumerated fields for positional arguments.
		pan.enum++
		pan.exists = true
		return true
	}

	return false
}

// Spec returns positional arguments specifier if given or empty string.
func (pan *posArgsNum) Spec() string {
	// The data in the array has the highest priority.
	switch {
	case pan.array != 0:
		return fmt.Sprintf("a1, ..., a%d", pan.array)
	case pan.enum != 0:
		return fmt.Sprintf("a1, ..., a%d", pan.enum)
	case pan.exists:
		return "args..."
	}

	return ""
}

// The getHelp returns automatically generated help information.
func getHelp(rv reflect.Value, f []*fieldSample, opts optSamples) (r string) {
	// If objects implements Helper interface try to calling
	// a custom Help method.
	if rv.Type().Implements(reflect.TypeOf((*Helper)(nil)).Elem()) {
		if m := rv.MethodByName("HelpOPT"); m.IsValid() {
			tmp := m.Call([]reflect.Value{reflect.ValueOf(opts["0"])})
			value := tmp[0].Interface()
			r = fmt.Sprintf("%s\n\n", value.(string))
		}
	}

	// If objects implements Usager interface try to calling
	// a custom Usage method.
	if rv.Type().Implements(reflect.TypeOf((*Usager)(nil)).Elem()) {
		if m := rv.MethodByName("UsageOPT"); m.IsValid() {
			tmp := m.Call([]reflect.Value{reflect.ValueOf(opts["0"])})
			value := tmp[0].Interface()
			r = fmt.Sprintf("%s%s\n", r, value.(string))
		}
	} else {
		r = fmt.Sprintf("%s%s", r, getUsageHelp(f, opts))
	}
	r = fmt.Sprintf("%s%s\n", r, getArgumentsHelp(f))

	return
}

// The combineNames correctly combines the short and long option name.
// If first is true returns the first non-empty name only.
func combineNames(short, long string, first bool) (r string) {
	var tmp string

	if short != "" {
		r = "-" + short
	}

	if (r == "" || !first) && long != "" {
		tmp = "--" + long
		if r != "" {
			r = fmt.Sprintf("%s,%s", r, tmp)
		} else {
			r = tmp
		}
	}

	return
}

// The cutHelp cut the help line on short pices.
// It is assumed that the maximum length of the left + right side should
// not exceed 79 characters.
func cutHelp(help string, start int) string {
	var (
		tmp    string
		lines  []string = []string{}
		indent string   = fmt.Sprintf(fmt.Sprintf("%s-%ds", "%", start), "")
	)

	tmp = indent
	for _, item := range strings.Split(help, " ") {
		if len(tmp+" "+item) <= 79 {
			tmp += " " + item
		} else {
			if tmp != indent {
				lines = append(lines, tmp)
				tmp = fmt.Sprintf("%5s%s %s", " ", indent, item)
			} else {
				lines = append(lines, " "+item)
				tmp = indent
			}
		}
	}

	lines = append(lines, tmp)
	return strings.TrimPrefix(strings.Join(lines, "\n"), indent)
}

// The getUsageHelp returns help in a single line -
// application launch structure.
func getUsageHelp(fields []*fieldSample, opts optSamples) (r string) {
	var (
		name string
		pan  = posArgsNum{}
	)

	// Get name of the application.
	path := opts["0"]
	_, file := filepath.Split(path)
	r = fmt.Sprintf("Usage: ./%s ", file)

	// Regular options.
	for _, f := range fields {
		// Ignore positional arguments but control
		// the number of required arguments.
		if c := pan.Count(f); c {
			continue
		}

		// Create launch options.
		name = combineNames(f.TagSample.Short, f.TagSample.Long, true)
		if f.Item.Kind() != reflect.Bool {
			name += " value"
		}

		r += fmt.Sprintf("[%s] ", name)
	}

	// Positional arguments.
	if spec := pan.Spec(); len(spec) != 0 {
		r += "-- " + spec
	}

	return
}

// The getArgumentsHelp returns a list of arguments
// as string and their help line.
func getArgumentsHelp(fields []*fieldSample) (r string) {
	var (
		leftside  string
		rightside string
		pattern   string
		lines     [][]string
		max       int

		pan = posArgsNum{}
	)

	// Regular options.
	for _, f := range fields {
		// Ignore positional arguments but control
		// the number of required arguments.
		if c := pan.Count(f); c {
			continue
		}

		// Determine available option names and create pattern for left side.
		leftside = combineNames(f.TagSample.Short, f.TagSample.Long, false)
		if m := len(leftside); m > max {
			max = m
			pattern = fmt.Sprintf("%s-%ds", "%", max+2)
		}

		// Help text.
		rightside = "..."
		if f.TagSample.Help != "" {
			rightside = f.TagSample.Help
		}

		// Add data into line.
		lines = append(lines, []string{leftside, rightside})
	}

	// Positional arguments.
	if spec := pan.Spec(); len(spec) != 0 {
		leftside = spec
		rightside = "positional arguments"
		if m := len(leftside); m > max {
			pattern = fmt.Sprintf("%s-%ds", "%", max+2)
		}
		lines = append(lines, []string{leftside, rightside})
	}

	// Join lines.
	for _, line := range lines {
		leftside = fmt.Sprintf(pattern, line[0])
		tmp := fmt.Sprintf("    %s%s", leftside, cutHelp(line[1], max+1))
		r += "\n" + tmp
	}

	return
}
