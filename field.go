package opt

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/goloop/scs"
)

const (
	// The tagNameOption the identifier of the tag that sets
	// the main option name.
	tagNameOpt = "opt"

	// The tagNameAlt the identifier of the tag that sets
	// the alternative option name.
	tagNameAlt = "alt"

	// The tagNameDefValue the identifier of the tag that
	// sets the default value.
	tagNameDefValue = "def"

	// The tagNameHelp the identifier of the tag that sets the help value.
	tagNameHelpMsg = "help"

	// The tagNameSepList the identifier of the tag that sets the delimiter
	// for the tagNameDefValue field if the struct field is a list.
	tagNameSepList = "sep"

	// The defValueIgnored is the value of the tagNameOption field that
	// should be ignored during processing.
	defValueIgnored = "-"

	// The defSep sets the default delimiter for the tagNameDefValue field
	// if a struct field is a list. An empty value indicates that the data
	// will not be divided into items, but will be considered as one item
	// in the list.
	defSep = ""
)

// isASCIILetter reports whether b is an ASCII letter (A-Z or a-z).
func isASCIILetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

// isOrderFlag reports whether s is a positional-index flag: a non-empty
// run of decimal digits ("0", "1", ..., "10", ...).
func isOrderFlag(s string) bool {
	if s == "" {
		return false
	}

	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

// isShortFlagSafe reports whether s is a single ASCII letter, i.e. a short
// option name without the special "?" and "[]" markers.
func isShortFlagSafe(s string) bool {
	return len(s) == 1 && isASCIILetter(s[0])
}

// isShortFlag reports whether s is a valid short option name: a single ASCII
// letter or one of the special markers "?" (doc field) and "[]" (positional).
func isShortFlag(s string) bool {
	return s == "?" || s == "[]" || isShortFlagSafe(s)
}

// isLongFlag reports whether s is a valid long option name: an ASCII letter
// followed by one or more letters, digits or dashes (at least two chars).
func isLongFlag(s string) bool {
	if len(s) < 2 || !isASCIILetter(s[0]) {
		return false
	}

	for i := 1; i < len(s); i++ {
		b := s[i]
		if !isASCIILetter(b) && (b < '0' || b > '9') && b != '-' {
			return false
		}
	}

	return true
}

// The tagGroup is the tag group of a field.
type tagGroup struct {
	shortFlag string // short flag
	longFlag  string // long flag
	defValue  string // default value
	helpMsg   string // help information about field
	sepList   string // list delimiter for defValue
	isIgnored bool   // true if ignore the field
}

// The fieldCast is field data structure.
type fieldCast struct {
	fieldName string         // just field name
	tagGroup  *tagGroup      // tga group
	item      *reflect.Value // field instance
}

// The fieldCastList is list of field data structure.
type fieldCastList []*fieldCast

// The flags function returns map of field's flags in opt and alt tags.
func (fcl fieldCastList) flags() map[string]int {
	result := make(map[string]int, len(fcl))

	for _, fc := range fcl {
		if fc.tagGroup.shortFlag != "" {
			result[fc.tagGroup.shortFlag]++
		}

		if fc.tagGroup.longFlag != "" {
			result[fc.tagGroup.longFlag]++
		}
	}

	return result
}

// The boolFlags returns the set of flag names (short and long) whose field is
// of bool kind. It lets the parser avoid consuming a following positional
// argument as the value of a boolean option.
func (fcl fieldCastList) boolFlags() map[string]bool {
	result := make(map[string]bool, len(fcl))

	for _, fc := range fcl {
		if fc.item == nil || fc.item.Kind() != reflect.Bool {
			continue
		}

		if fc.tagGroup.shortFlag != "" {
			result[fc.tagGroup.shortFlag] = true
		}

		if fc.tagGroup.longFlag != "" {
			result[fc.tagGroup.longFlag] = true
		}
	}

	return result
}

// The validateStruct checks whether the object is a pointer to the structure,
// and returns reflect.Type and reflect.Value of the object. If the object is
// not a pointer to the structure or object is nil, it returns an error.
func validateStruct(obj any) (reflect.Type, reflect.Value, error) {
	rt, rv, err := reflect.TypeOf(obj), reflect.ValueOf(obj), error(nil)

	// Check object type
	// Object should be a pointer to a non-empty struct.
	if obj == nil {
		err = errors.New("obj is nil")
	} else if rv.Kind() != reflect.Ptr || rv.IsNil() {
		err = errors.New("obj should be a non-nil pointer to a struct")
	} else if rv.Type().Elem().Kind() != reflect.Struct {
		err = errors.New("obj should be a pointer to a struct")
	} else if rv.Elem().NumField() == 0 {
		err = errors.New("obj should be a pointer to a non-empty struct")
	}

	return rt, rv, err
}

// The getFieldCastList parses the structure fields and
// returns list of the fieldCast.
func getFieldCastList(obj any) (fieldCastList, error) {
	var result fieldCastList

	// Check object type.
	rt, rv, err := validateStruct(obj)
	if err != nil {
		return result, err
	}

	elem := rv.Elem()
	urlS, urlP := reflect.TypeOf(url.URL{}), reflect.TypeOf((*url.URL)(nil))
	for i := 0; i < elem.NumField(); i++ {
		// Get tag data from the field.
		field := rt.Elem().Field(i)

		// Get tag group.
		tg, err := getTagGroup(
			field.Name,
			strings.Trim(field.Tag.Get(tagNameOpt), " -"),
			strings.Trim(field.Tag.Get(tagNameAlt), " -"),
			field.Tag.Get(tagNameDefValue),
			field.Tag.Get(tagNameSepList),
			field.Tag.Get(tagNameHelpMsg),
		)

		if err != nil {
			return result, err
		} else if tg.isIgnored {
			continue
		}

		// Collect fields for further analysis.
		item := elem.FieldByName(field.Name)
		fc := fieldCast{fieldName: field.Name, tagGroup: &tg, item: &item}

		kind := fc.item.Kind()
		switch f := fc.tagGroup.shortFlag; {
		case f == "?" && kind != reflect.String:
			// To load doc, the field must be of the string type.
			err = fmt.Errorf("%s field should be a string", fc.fieldName)
		case f == "[]" && kind != reflect.Array && kind != reflect.Slice:
			// To load positional arguments,
			// the field must be of the slice type.
			err = fmt.Errorf("%s field should be a list", fc.fieldName)
		case kind == reflect.Struct && fc.item.Type() != urlS:
			// Supported url.URL struct only.
			err = fmt.Errorf("%s field has invalid type", fc.fieldName)
		case kind == reflect.Ptr:
			// Pointer to the structure *url.URL only.
			k, t := fc.item.Type().Elem().Kind(), fc.item.Type()
			if k == reflect.Struct && t != urlP {
				err = fmt.Errorf("%s field has invalid type", fc.fieldName)
			}
		}

		if err != nil {
			return result, err
		}

		result = append(result, &fc)
	}

	return result, nil
}

// getTagGroup returns a tagGroup with the specified tag values.
func getTagGroup(
	fieldName,
	optTagValue,
	altTagValue,
	defTagValue,
	sepTagValue,
	helpTagValue string,
) (tagGroup, error) {
	var err error
	// Create tag group with default values.
	tg := tagGroup{
		defValue: defTagValue,
		sepList:  sepTagValue,
		helpMsg:  helpTagValue,
	}

	// The fieldName must be used for an empty value of optTagValue.
	if optTagValue == "" {
		// Convert pascal case to kebab case for long flag name only.
		optTagValue = fieldName
		if utf8.RuneCountInString(fieldName) > 1 {
			if kebab, err := scs.PascalToKebab(fieldName); err == nil {
				optTagValue = kebab
			}
		}
	}

	msg := "invalid %s tag value %s"
	if isOrderFlag(optTagValue) || isShortFlag(optTagValue) {
		tg.shortFlag = optTagValue
		if isLongFlag(altTagValue) {
			tg.longFlag = altTagValue
		} else if altTagValue != "" {
			err = fmt.Errorf(msg, tagNameAlt, altTagValue)
		}
	} else if isLongFlag(optTagValue) {
		tg.longFlag = optTagValue
		if isShortFlagSafe(altTagValue) {
			tg.shortFlag = altTagValue
		} else if altTagValue != "" {
			err = fmt.Errorf(msg, tagNameAlt, altTagValue)
		}
	} else if optTagValue == defValueIgnored {
		tg.isIgnored = true
	} else {
		err = fmt.Errorf(msg, tagNameOpt, optTagValue)
	}

	tg.longFlag = strings.ToLower(tg.longFlag)
	if utf8.RuneCountInString(tg.longFlag) > 32 {
		err = fmt.Errorf("%s is a very long name, max 32 chars", tg.longFlag)
	}

	return tg, err
}
