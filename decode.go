package opt

import (
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// unmarshalOpt parses variables from the command-line and sets them into
// fields of object.
//
// Returns an error if it is impossible to parse the command line, for example:
// there are an arg not provided in the go-structure; several values will be
// passed for a field that is not a slice or array; etc.
//
// Generates panic if the structure has fields of the wrong type, for example:
// field with the "?" key is not a string; the field with the "[]" key is not
// a slice or array; the object is not a pointer; for unsupported field types
// like: chan, map, etc..
//
// unmarshalOpt method supports the following field's types: int, int8, int16,
// int32, int64, uin, uint8, uin16, uint32, in64, float32, float64, string,
// bool, url.URL and pointers, array or slice from thous types (i.e. *int, ...,
// []int, ..., []bool, ..., [2]*url.URL, etc.).
func unmarshalOpt(obj interface{}, args []string) error {
	// Analyze the structure and return a list of molds
	// of each field: field name, tag group and field pointer.
	//
	// If it's returns an error - be critical!
	// This is a problem with an incorrect structure and we need to
	// stop the program until the developer fixes this error.
	fcl, err := getFieldCastList(obj)
	if err != nil {
		// Convert this error to panic because it's a problem of
		// structure's fields (it's developer's problem).
		panic(err)
	}

	// Parse options.
	am := argMap{}
	if err := am.parse(args, fcl.flags()); err != nil {
		return err
	}

	// Insert values into the fields of the structure
	// from the command line arguments.
	for _, fc := range fcl {
		var err error

		if fc.tagGroup.shortFlag == "?" {
			// Generate help info.
			// The field must be of the string type, see in
			// the getFieldCastList function.
			help := getHelp(fcl, am)
			fc.item.Set(reflect.ValueOf(help))
			continue
		}

		// Contains a couple of arguments like: U, users, U or/and users -
		// where U and users is synonyms.
		arg := fmt.Sprintf(
			"%s tmp/and %s",
			fc.tagGroup.shortFlag,
			fc.tagGroup.longFlag,
		)
		arg = strings.Trim(arg, " or/and ")

		value, kind := []string{}, fc.item.Kind()
		switch f := fc.tagGroup.shortFlag; {
		case f == "[]":
			// Get positional arguments.
			value = am.posValues()
		default:
			// Get the values of the argument.
			value = am.flagValue(
				fc.tagGroup.shortFlag,
				fc.tagGroup.longFlag,
				fc.tagGroup.defValue,
				fc.tagGroup.sepList,
			)

			// The user in the command line tries to pass arguments as
			// list to a field that doesn't have the slice or array type.
			if len(value) > 1 {
				if kind != reflect.Array && kind != reflect.Slice {
					// In this situation, we need to take the
					// last value in the list.
					//
					// return fmt.Errorf("%s used more than once", arg)
					value = []string{value[len(value)-1]}
				}
			}
		}

		// Set values of the desired type.
		switch kind {
		case reflect.Array:
			// If a separator is specified, the elements must be separated.
			var result []string

			if sep := fc.tagGroup.sepList; sep != "" {
				for _, item := range value {
					tmp := strings.Split(item, sep)
					result = append(result, tmp...)
				}
			} else {
				result = value
			}

			if max := fc.item.Type().Len(); len(result) > max {
				// Array overflow.
				// -> "%d items overflow [%d]%v array", len(result), max, kind,
				// kind := fs.item.Index(0).Kind()
				return fmt.Errorf(
					"maximum number of values for %s argument "+
						"is %d but passed %d values",
					arg, max, len(result),
				)
			}

			err = setSequence(fc.item, result)
		case reflect.Slice:
			// Be sure to set Len equal Cap and more than zero.
			// The slice must have at least one element to determine
			// the type of the one.
			// If a separator is specified, the elements must be separated.
			var result []string

			if sep := fc.tagGroup.sepList; sep != "" {
				for _, item := range value {
					tmp := strings.Split(item, sep)
					result = append(result, tmp...)
				}
			} else {
				result = value
			}

			if len(result) != 0 {
				size := len(result)
				tmp := reflect.MakeSlice(fc.item.Type(), size, size)
				err = setSequence(&tmp, result)
				if err == nil {
					fc.item.Set(reflect.AppendSlice(*fc.item, tmp))
				}
			}
		case reflect.Ptr:
			if fc.item.Type().Elem().Kind() != reflect.Struct {
				// If the pointer is not to a structure.
				tmp := reflect.Indirect(*fc.item)
				err = setValue(tmp, value[len(value)-1])
			} else {
				// If a pointer to a structure of the url.URL.
				err = setValue(*fc.item, value[len(value)-1])
			}
		case reflect.Struct:
			// Structure of the url.URL.
			err = setValue(*fc.item, value[len(value)-1])
		default:
			// Set any type.
			err = setValue(*fc.item, value[len(value)-1])
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// The setSequence sets slice into item.
func setSequence(item *reflect.Value, seq []string) (err error) {
	// defer func() {
	// 	// Catch the panic and return an exception as a value.
	// 	if r := recover(); r != nil {
	// 		err = fmt.Errorf("%v", r)
	// 	}
	// }()

	if item.Len() != len(seq) {
		return nil
	}

	// Set values from sequence.
	for i, value := range seq {
		elem := item.Index(i)
		err := setValue(elem, value)
		if err != nil {
			return err
		}
	}

	return nil
}

// The setValue sets value into field.
func setValue(item reflect.Value, value string) (err error) {
	defer func() {
		// Catch the panic and return an exception as a value.
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	var kind = item.Kind()

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		r, err := strToIntKind(value, kind)
		if err != nil {
			return err
		}
		item.SetInt(r)
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		r, err := strToUintKind(value, kind)
		if err != nil {
			return err
		}
		item.SetUint(r)
	case reflect.Float32, reflect.Float64:
		r, err := strToFloatKind(value, kind)
		if err != nil {
			return err
		}
		item.SetFloat(r)
	case reflect.Bool:
		r, err := strToBool(value)
		if err != nil {
			return err
		}
		item.SetBool(r)
	case reflect.String:
		item.SetString(value)
	case reflect.Ptr:
		// // Will not be allowed by the getFieldCast method.
		// if item.Type() != reflect.TypeOf((*url.URL)(nil)) {
		// 	return fmt.Errorf("%s field has invalid type", item.Type().Name())
		// }

		// The url.URL struct only.
		u, err := url.Parse(value)
		if err != nil {
			return err
		}

		item.Set(reflect.ValueOf(u))
	case reflect.Struct:
		// // Will not be allowed by the getFieldCast method.
		// if item.Type() != reflect.TypeOf(url.URL{}) {
		// 	return fmt.Errorf("%s field has invalid type", item.Type().Name())
		// }

		// The url.URL struct only.
		u, err := url.Parse(value)
		if err != nil {
			return err
		}

		item.Set(reflect.ValueOf(*u))
	default:
		return fmt.Errorf("%s field has invalid type", item.Type().Name())
	}

	return nil
}

// The strToIntKind convert string to int64 type with checking for conversion
// to intX type. Returns default value for int type if value is empty.
//
// P.s. The intX determined by reflect.Kind.
func strToIntKind(value string, kind reflect.Kind) (r int64, err error) {
	// For empty string returns zero.
	if len(value) == 0 {
		return 0, nil
	}

	// Convert string to int64.
	r, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("'%v' is incorrect value", value)
	}

	switch kind {
	case reflect.Int:
		// If there was no exception during the conversion,
		// then we have exactly the number in the uint64 range, but
		// for 32-bit platform it is necessary to check overflow.
		if strconv.IntSize == 32 {
			if r < math.MinInt32 || r > math.MaxInt32 {
				return 0, fmt.Errorf("%d overflows int32", r)
			}
		}
	case reflect.Int8:
		if r < math.MinInt8 || r > math.MaxInt8 {
			return 0, fmt.Errorf("%d overflows int8", r)
		}
	case reflect.Int16:
		if r < math.MinInt16 || r > math.MaxInt16 {
			return 0, fmt.Errorf("%d overflows int16", r)
		}
	case reflect.Int32:
		if r < math.MinInt32 || r > math.MaxInt32 {
			return 0, fmt.Errorf("%d overflows int32", r)
		}
	case reflect.Int64:
		// If there was no exception during the conversion,
		// then we have exactly the number in the int64 range.
	default:
		r, err = 0, fmt.Errorf("incorrect kind %v", kind)
	}

	return
}

// strToUintKind convert string to uint64 type with checking for conversion
// to uintX type. Returns default value for uint type if value is empty.
//
// P.s. The uintX determined by reflect.Kind.
func strToUintKind(value string, kind reflect.Kind) (r uint64, err error) {
	// For empty string returns zero.
	if len(value) == 0 {
		return 0, nil
	}

	// Convert string to uint64.
	r, err = strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(
			"'%v' has incorrect type, positive number expected",
			value,
		)
	}

	switch kind {
	case reflect.Uint:
		// If there was no exception during the conversion,
		// then we have exactly the number in the uint64 range, but
		// for 32-bit platform it is necessary to check overflow.
		if strconv.IntSize == 32 && r > math.MaxUint32 {
			return 0, fmt.Errorf("%d overflows uint32", r)
		}
	case reflect.Uint8:
		if r > math.MaxUint8 {
			return 0, fmt.Errorf("%d overflows uint8", r)
		}
	case reflect.Uint16:
		if r > math.MaxUint16 {
			return 0, fmt.Errorf("%d overflows uint16", r)
		}
	case reflect.Uint32:
		if r > math.MaxUint32 {
			return 0, fmt.Errorf("strToUint32: %d overflows uint32", r)
		}
	case reflect.Uint64:
		// If there was no exception during the conversion,
		// then we have exactly the number in the uint64 range.
	default:
		r, err = 0, fmt.Errorf("incorrect kind %v", kind)
	}

	return
}

// strToFloatKind convert string to float64 type with checking for conversion
// to floatX type. Returns default value for float64 type if value is empty.
//
// P.s. The floatX determined by reflect.Kind.
func strToFloatKind(value string, kind reflect.Kind) (r float64, err error) {
	// For empty string returns zero.
	if len(value) == 0 {
		return 0.0, nil
	}

	// Convert string to Float64.
	r, err = strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0, fmt.Errorf(
			"'%v' has incorrect type, number expected",
			value,
		)
	}

	switch kind {
	case reflect.Float32:
		if math.Abs(r) > math.MaxFloat32 {
			return 0.0, fmt.Errorf("%f overflows float32", r)
		}
	case reflect.Float64:
		// If there was no exception during the conversion,
		// then we have exactly the number in the float64 range.
	default:
		r, err = 0, fmt.Errorf("incorrect kind")
	}

	return
}

// strToBool convert string to bool type. Returns: result, error.
// Returns default value for bool type if value is empty.
func strToBool(value string) (bool, error) {
	var epsilon = math.Nextafter(1, 2) - 1

	// For empty string returns false.
	if len(value) == 0 {
		return false, nil
	}

	r, errB := strconv.ParseBool(value)
	if errB != nil {
		f, errF := strconv.ParseFloat(value, 64)
		if errF != nil {
			return r, fmt.Errorf(
				"'%v' has incorrect type, bool expected",
				value,
			)
		}

		if math.Abs(f) > epsilon {
			r = true
		}
	}

	return r, nil
}
