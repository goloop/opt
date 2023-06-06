package opt

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// The following concepts are used to generate text help
// based on command line arguments:
//
// |******| _________________________________________________________ Title
// Options:           |*****************************************| ___ Suffix
//     -v, --verbose - when set to true, displays more detailed \
//                     information about the job, otherwise only| --- Item
//                     critical information will be displayed;  /
//         --debug   - when set to true, displays log info.
//                  |*| _____________________________________________ Separator
// |****************| _______________________________________________ Prefix

// The separator separates the command line (prefix)
// from the documentation line (suffix).
const separator = " "

// The optionItems is struct of the help line for options section.
type optionItems struct {
	short string // short flag name
	long  string // long flag name
	items string // prefix of items
	help  string // information string
}

// The posItems is struct of the help line for positional section.
type posItems struct {
	short string // short flag name: 1, 2, ..., N
	help  string // information string
}

// The getOptionPrefix returns the prefix of the documentation line in which
// the options are specified: -v; -v, --verbose; --verbose. The second
// argument is the length of this one.
func getOptionPrefix(short, long string) (string, int) {
	var result string

	// Add one dash to the short flag and comma if there is a long flag.
	if short != "" {
		short = fmt.Sprintf("-%s", short)
		if long != "" {
			short = fmt.Sprintf("%s, ", short)
		}
	}

	// Add two dashes to the long flag.
	if long != "" {
		long = fmt.Sprintf("--%s", long)
	}

	// The layout has a format: 4 spaces, 4 positions for a short flag,
	// the everything else is a long flag.
	result = strings.TrimRight(fmt.Sprintf("    %-4s%s", short, long), " ")
	return result, utf8.RuneCountInString(result)
}

// The wrapHelpMsg splits the line doc by the right extreme space,
// into lines no more than wc in length, taking into account the
// length of the tab.
//
// Adds a sep to the first line.
func wrapHelpMsg(sep, str string, tab, wc int) []string {
	var result []string

	// Don't add a prefix if help string isn't specified.
	if str == "" {
		return result
	}

	// Line wrapping occurs by words. Divide the line into words and
	// create new lines where the length of the words does not exceed
	// the specified.
	wc = wc - tab // line length without tabs
	line, count := "", utf8.RuneCountInString
	for _, word := range strings.Split(str, " ") {
		l := count(line) + count(word)

		switch {
		case line == "":
			line = word
		case l >= wc:
			if len(result) == 0 {
				line = sep + line
			}
			result = append(result, line)
			line = word
		default:
			line += " " + word
		}
	}

	if line != "" {
		if len(result) == 0 {
			line = sep + line
		}
		result = append(result, line)
	}

	return result
}

// The getOptionBlock returns the text of the documentation
// for optional arguments. The second parameter will damage -1 if
// there is no container `[]` for processing positional parameters,
// 0 - if such a container exists and its size is not limited (for slice),
// more than 0 if the number of elements in this container is limited (array).
func getOptionBlock(fcl fieldCastList, am argMap) (string, int) {
	lines := []string{}

	// Go through all the fields, make a prefix for the help line,
	// which includes the available arguments. Determine the largest
	// prefix of arguments. Ignore fields that do not require a optionItems.
	maxPrefixLen := 0      // the maximum length of the prefix
	posArgsExists := false // true if some field has [] opt marker
	posArgsLen := 0        // number of positional arguments
	items := make([]optionItems, 0, len(fcl))
	for _, fc := range fcl {
		// Ignore technical fields.
		switch flag := fc.tagGroup.shortFlag; {
		case fc.tagGroup.shortFlag == "[]":
			// The number of positional arguments in an array is
			// limited by its size. For slice are no restrictions.
			switch fc.item.Kind() {
			case reflect.Array:
				posArgsLen = fc.item.Len()
				if posArgsLen > 0 {
					posArgsLen-- // 0 element it's an app ptah
				}
				posArgsExists = posArgsLen != 0
			case reflect.Slice:
				posArgsLen = 0
				posArgsExists = true
			}
			fallthrough
		case fc.tagGroup.shortFlag == "?":
			// Field for uploading documentation.
			fallthrough
		case orderFlagRgx.Match([]byte(flag)):
			// Fixed positional argument.
			continue
		}

		// Make prefix from the items.
		p, l := getOptionPrefix(fc.tagGroup.shortFlag, fc.tagGroup.longFlag)
		items = append(items, optionItems{
			fc.tagGroup.shortFlag,
			fc.tagGroup.longFlag,
			p,
			fc.tagGroup.helpMsg,
		})

		// Determine the largest prefix of arguments. The option is doesn't
		// displayed in the documentation if it doesn't have a help message.
		if l > maxPrefixLen && fc.tagGroup.helpMsg != "" {
			maxPrefixLen = l
		}
	}

	// Add title "Options" if this part is exists.
	if len(items) != 0 {
		lines = append(lines, "Options:")
		sort.Slice(items, func(i, j int) bool {
			return items[i].short < items[j].short &&
				items[i].long < items[j].long
		})
	}

	// Concatenation of argument prefix and suffix.
	// Such a line cannot be too long.
	sep, rcis := separator, utf8.RuneCountInString
	for _, item := range items {
		// Add a semicolon to each line, and if this
		// is the last line, a period.
		help := item.help
		switch {
		case help == "":
			continue
		default:
			help += ";"
		}

		for j, l := range wrapHelpMsg(sep, help, maxPrefixLen, 79) {
			if j == 0 {
				tpl := fmt.Sprintf("%%-%ds%%s", maxPrefixLen)
				lines = append(lines, fmt.Sprintf(tpl, item.items, l))
				continue
			}

			tpl := fmt.Sprintf("%%%ds", maxPrefixLen+rcis(l)+len(sep))
			lines = append(lines, fmt.Sprintf(tpl, l))
		}

	}

	// Result.
	result := ""
	if top := len(lines); top != 0 {
		lines[top-1] = strings.TrimSuffix(lines[top-1], ";") + "."
		result = strings.Join(lines, "\n")
	}

	pos := -1
	if posArgsExists {
		pos = posArgsLen
	}

	return result, pos
}

// The getPositionalBlock returns the text of the
// documentation about positional arguments.
func getPositionalBlock(fcl fieldCastList, posArgsLen int) string {
	var lines []string

	// Collect positional arguments.
	items := make([]posItems, 0, len(fcl))
	for _, fc := range fcl {
		if f := fc.tagGroup.shortFlag; orderFlagRgx.Match([]byte(f)) {
			items = append(items, posItems{f, fc.tagGroup.helpMsg})
		}
	}

	// Sort position arguments.
	sort.Slice(items, func(i, j int) bool {
		return items[i].short < items[j].short
	})

	// Update the count of positional arguments.
	// P.s. use len(items)-1 because the null argument doesn't
	// count as it is the path to the application.
	if l := len(items) - 1; posArgsLen <= 0 && l >= 0 {
		id, err := strconv.Atoi(items[l].short)
		switch {
		case err != nil:
			fallthrough
		case id <= l:
			posArgsLen = l
		default:
			posArgsLen = id
		}
	}

	// Create information.
	maxPrefixLen, subitems := 6, []string{}
	sep, rcis := separator, utf8.RuneCountInString
	for _, item := range items {
		// Add a semicolon to each line, and if this
		// is the last line, a period.
		help := item.help
		switch {
		case help == "" || item.short == "0":
			continue
		default:
			help += ";"
		}

		for j, l := range wrapHelpMsg(sep, help, maxPrefixLen, 79) {
			if j == 0 {
				tpl := fmt.Sprintf("%%%ds%%s", maxPrefixLen)
				subitems = append(subitems, fmt.Sprintf(tpl, item.short, l))
				continue
			}

			tpl := fmt.Sprintf("%%%ds", maxPrefixLen+rcis(l)+len(sep))
			subitems = append(subitems, fmt.Sprintf(tpl, l))
		}
	}

	// Create subtitle.
	subtitle := ""
	switch {
	case posArgsLen < 0:
		return ""
	case posArgsLen == 0:
		subtitle = "The app takes an unlimited number of positional arguments"
	case posArgsLen == 1:
		subtitle = "The app takes an one of positional argument"
	default:
		tpl := "The app takes %d of the positional arguments"
		subtitle = fmt.Sprintf(tpl, posArgsLen)
	}

	if top := len(subitems); top != 0 {
		subitems[top-1] = strings.TrimSuffix(subitems[top-1], ";") + "."
		subtitle += ", including:"
	} else {
		subtitle += "."
	}

	// Result.
	lines = append(lines, "Positional arguments:")
	lines = append(lines, strings.Join(wrapHelpMsg("", subtitle, 0, 79), "\n"))
	if len(items) != 0 {
		lines = append(lines, subitems...)
	}

	return strings.Join(lines, "\n")
}

// The getHelp returns help on using command line options.
func getHelp(fcl fieldCastList, am argMap) string {
	var result []string

	// Generate option block.
	optText, posArgsLen := getOptionBlock(fcl, am)
	if optText != "" {
		result = append(result, optText)
	}

	// Generate positional block.
	posText := getPositionalBlock(fcl, posArgsLen)
	if posText != "" {
		if optText != "" {
			result = append(result, "")
		}
		result = append(result, posText)
	}

	return strings.Join(result, "\n")
}
