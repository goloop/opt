package opt

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
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

var (
	// The orderFlagRgx a regular expression to check if a string is a number.
	orderFlagRgx = regexp.MustCompile(`^\d+$`)

	// The shortFlagRgx a regular expression to check
	// if a string is short option.
	shortFlagRgx = regexp.MustCompile(`^(\?|\[\]|[A-Za-z]{1})$`)

	// The shortFlagSafeRgx as the shortFlagRgx but without the
	// ability to win special tags like: ?, [].
	shortFlagSafeRgx = regexp.MustCompile(`^([A-Za-z]{1})$`)

	// The longRgx a regular expression to check if a string is long option.
	longFlagRgx = regexp.MustCompile(`^[A-Za-z]{1}[A-Za-z\-1-9]{1,}$`)
)

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

// The flags returns map of field's flags in opt and alt tags.
func (fcl fieldCastList) flags() map[string]int {
	var result = make(map[string]int, len(fcl))

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

// The getFieldCastList parses the structure fields and
// returns list of the fieldCast.
func getFieldCastList(obj interface{}) (fieldCastList, error) {
	var result = fieldCastList{}

	// The obj argument should be a pointer to initialized object.
	rt, rv := reflect.TypeOf(obj), reflect.ValueOf(obj)
	switch {
	case obj == nil:
		fallthrough
	case rt.Kind() != reflect.Ptr:
		fallthrough
	case rt.Elem().Kind() != reflect.Struct:
		fallthrough
	case !rv.Elem().IsValid():
		err := errors.New("obj should be a pointer to an initialized struct")
		return result, err
	}

	elem := rv.Elem()
	for i := 0; i < elem.NumField(); i++ {
		var err error

		// Get tag data from the field.
		field := rt.Elem().Field(i)

		// Get tag group.
		help := field.Tag.Get(tagNameHelpMsg)
		def := field.Tag.Get(tagNameDefValue)
		alt := strings.Trim(field.Tag.Get(tagNameAlt), " -")

		sep := defSep
		if s := field.Tag.Get(tagNameSepList); s != "" {
			sep = s
		}

		opt := field.Tag.Get(tagNameOpt)
		if opt != defValueIgnored {
			// Clear only if this field is not ignored.
			strings.Trim(field.Tag.Get(tagNameOpt), " -")
		}

		tg, err := getTagGroup(field.Name, opt, alt, def, sep, help)
		switch {
		case err != nil:
			return result, err
		case tg.isIgnored:
			continue
		}

		// Collect fields for further analysis.
		item := elem.FieldByName(field.Name)
		fc := fieldCast{field.Name, &tg, &item}

		// ...
		kind := fc.item.Kind()
		switch f := fc.tagGroup.shortFlag; {
		case f == "?" && kind != reflect.String:
			// To load doc, the field must be of the string type.
			err = fmt.Errorf("%s field should be a string", fc.fieldName)
		case f == "[]" && kind != reflect.Array && kind != reflect.Slice:
			// To load positional arguments,
			// the field must be of the slice type.
			err = fmt.Errorf("%s field should be a list", fc.fieldName)
		case kind == reflect.Struct:
			// Supported url.URL struct only.
			if fc.item.Type() != reflect.TypeOf(url.URL{}) {
				err = fmt.Errorf("%s field has invalid type", fc.fieldName)
			}
		case kind == reflect.Ptr:
			// Pointer to the structure *url.URL only.
			k, t := fc.item.Type().Elem().Kind(), fc.item.Type()
			if k == reflect.Struct && t != reflect.TypeOf((*url.URL)(nil)) {
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

// The getTagGroup ...
func getTagGroup(
	fieldName string,
	optTagValue string,
	altTagValue string,
	defTagValue string,
	sepTagValue string,
	helpTagValue string,
) (tagGroup, error) {
	var tg = tagGroup{}

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

	// Default value and help info.
	tg.defValue = defTagValue
	tg.sepList = sepTagValue
	tg.helpMsg = helpTagValue

	switch optByte, altByte := []byte(optTagValue), []byte(altTagValue); {
	case orderFlagRgx.Match(optByte):
		tg.shortFlag = optTagValue
	case shortFlagRgx.Match(optByte):
		tg.shortFlag = optTagValue
		if longFlagRgx.Match(altByte) {
			tg.longFlag = altTagValue
		} else if len(altTagValue) != 0 {
			return tg, fmt.Errorf(
				"invalid %s tag value %s",
				tagNameAlt, altTagValue,
			)
		}
	case longFlagRgx.Match(optByte):
		tg.longFlag = optTagValue
		if shortFlagSafeRgx.Match(altByte) {
			tg.shortFlag = altTagValue
		} else if len(altTagValue) != 0 {
			return tg, fmt.Errorf(
				"invalid %s tag value %s",
				tagNameAlt, altTagValue,
			)
		}
	default:
		if optTagValue == defValueIgnored {
			tg.isIgnored = true
		} else {
			return tg, fmt.Errorf(
				"invalid %s tag value %s",
				tagNameOpt, optTagValue,
			)
		}
	}

	// The long flag is always lowercase.
	if tg.longFlag != "" {
		tg.longFlag = strings.ToLower(tg.longFlag)
	}

	// The maximum length of the flag cannot exceed 32 characters.
	if utf8.RuneCountInString(tg.longFlag) > 32 {
		return tg, fmt.Errorf(
			"%s is a very long name, maximum 32 characters",
			tg.longFlag,
		)
	}

	return tg, nil
}
