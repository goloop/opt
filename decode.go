package opt

import (
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// durationType is the reflect.Type of time.Duration, used to accept
// human-readable durations like "30s" or "1h30m" for such fields.
var durationType = reflect.TypeOf(time.Duration(0))

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
// int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string,
// bool, url.URL and pointers, array or slice from those types (i.e. *int, ...,
// []int, ..., []bool, ..., [2]*url.URL, etc.).
func unmarshalOpt(obj any, args []string) []error {
	var errs []error

	// A wrong obj - nil, not a pointer, or not a pointer to a non-empty struct -
	// is a runtime caller error (obj may arrive from a variable), so return it
	// instead of panicking, as documented and as encoding/json does. A bad
	// struct TAG remains a panic below: that is a developer error, visible on
	// the first run.
	if _, _, err := validateStruct(obj); err != nil {
		return []error{err}
	}

	// Analyze the structure and return a list of molds
	// of each field: field name, tag group and field pointer.
	//
	// A remaining error here is a tag/type misconfiguration of the struct, so
	// be critical and stop the program until the developer fixes it.
	fcl, err := getFieldCastList(obj)
	if err != nil {
		panic(err)
	}

	// Parse options.
	am := argMap{}
	if err := am.parse(args, fcl.flags(), fcl.boolFlags()); err != nil {
		errs = append(errs, err)
	}

	// Insert values into the fields of the structure
	// from the command line arguments.
	//
	// Important: Do not interrupt the cycle on the first error:
	//  - if there is a field for reference information,
	//    it should be processed in any case;
	//  - need to collect all possible errors.
	for _, fc := range fcl {
		if fc.tagGroup.shortFlag == "?" {
			// Store generated help info. The text depends only on the
			// struct type, so it is generated once per type and cached.
			// The field is a string (verified by getFieldCastList).
			fc.item.SetString(cachedHelp(reflect.TypeOf(obj).Elem(), fcl))
			continue
		}

		// Name of the option for error messages, e.g. "-p/--port".
		arg := optName(fc.tagGroup.shortFlag, fc.tagGroup.longFlag)

		value, kind, ok := []string{}, fc.item.Kind(), false
		isPositional := fc.tagGroup.shortFlag == "[]"
		switch {
		case isPositional:
			// Get positional arguments.
			value = am.posValues()
		default:
			// Get the values of the argument.
			value, ok = am.flagValue(
				fc.tagGroup.shortFlag,
				fc.tagGroup.longFlag,
				fc.tagGroup.defValue,
				fc.tagGroup.sepList,
			)

			// A repeated option was given to a non-list field: keep
			// the last value (last one wins).
			if len(value) > 1 {
				if kind != reflect.Array && kind != reflect.Slice {
					value = []string{value[len(value)-1]}
				}
			}
		}

		// A required option must be present on the command line. For a
		// positional list that means at least one positional argument.
		if fc.tagGroup.required {
			provided := ok || (isPositional && len(value) > 0)
			if !provided {
				errs = append(errs, fmt.Errorf("%s is required", arg))
				continue
			}
		}

		// An absent flag never clobbers a value the caller pre-filled,
		// mirroring how encoding/json leaves fields missing from the input
		// untouched. Without a def tag the field is skipped entirely; a def
		// applies only while the field still holds its zero value. This is
		// what makes the usual layering work: defaults, then environment,
		// then flags - each stage only overrides what it actually names.
		if !ok && !isPositional {
			if fc.tagGroup.defValue == "" || !fc.item.IsZero() {
				continue
			}
		}

		// If the argMap.flagValue hasn't value it's returns
		// []string{defValue} where defValue can be like "" - this is not
		// valid for a list because if the command line argument has no
		// data for the list this list must be empty!
		//
		// This heuristic applies only to flag-valued lists: a positional list
		// is fed by posValues(), where ok is always false and a real empty
		// positional argument ("") must be kept, not discarded.
		isEmptyList := !ok && !isPositional && len(value) == 1 && value[0] == ""

		// Set values of the desired type.
		switch kind {
		case reflect.Array:
			if isEmptyList {
				break
			}

			// If a separator is specified, the elements must be separated.
			result := splitValues(value, fc.tagGroup.sepList)
			if max := fc.item.Type().Len(); len(result) > max {
				// Array overflow: record the error and keep going so the
				// remaining fields (and any collected errors) are still
				// reported via errors.Join, as the package contract promises.
				errs = append(errs, fmt.Errorf(
					"maximum number of values for %s argument "+
						"is %d but passed %d values",
					arg, max, len(result),
				))
				continue
			}

			err = setSequence(fc.item, result)
		case reflect.Slice:
			if isEmptyList {
				break
			}

			// If a separator is specified, the elements must be separated.
			result := splitValues(value, fc.tagGroup.sepList)
			if len(result) != 0 {
				size := len(result)
				tmp := reflect.MakeSlice(fc.item.Type(), size, size)
				err = setSequence(&tmp, result)
				if err == nil {
					// Replace the slice contents, like encoding/json, so a
					// pre-populated field or a second Unmarshal does not
					// accumulate values.
					fc.item.Set(tmp)
				}
			}
		case reflect.Ptr:
			if fc.item.Type().Elem().Kind() != reflect.Struct {
				// A pointer to a scalar (e.g. *int). Leave a truly absent
				// optional pointer nil; otherwise allocate it so its element
				// can be set instead of dereferencing a nil pointer (which
				// previously failed on every call with a leaked reflect error).
				last := value[len(value)-1]
				if !ok && last == "" {
					break
				}
				if fc.item.IsNil() {
					fc.item.Set(reflect.New(fc.item.Type().Elem()))
				}
				err = setValue(reflect.Indirect(*fc.item), last)
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
			errs = append(errs, err)
		}
	}

	return errs
}

// optName formats an option's display name for error messages, e.g.
// "-p/--port", "--host", "-h", "argument #1" or "positional arguments".
func optName(short, long string) string {
	switch {
	case short == "[]":
		return "positional arguments"
	case isOrderFlag(short):
		return "argument #" + short
	case short != "" && long != "":
		return "-" + short + "/--" + long
	case long != "":
		return "--" + long
	case short != "":
		return "-" + short
	default:
		return ""
	}
}

// splitValues splits every element of values by sep and returns the
// flattened result. An empty sep leaves the values unchanged.
func splitValues(values []string, sep string) []string {
	if sep == "" {
		return values
	}

	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, strings.Split(v, sep)...)
	}

	return out
}

// The setSequence fills item (array or slice) from seq, left to right.
//
// For a fixed array, the size is the maximum number of elements: any
// remaining elements keep their zero value when fewer values are given.
// An oversized sequence is rejected earlier (array overflow), so seq is
// never longer than item here; the bound is kept only defensively.
func setSequence(item *reflect.Value, seq []string) error {
	n := len(seq)
	if l := item.Len(); n > l {
		n = l
	}

	for i := 0; i < n; i++ {
		if err := setValue(item.Index(i), seq[i]); err != nil {
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

	kind := item.Kind()

	// time.Duration is an int64; accept human-readable forms like "30s",
	// "1h30m". An empty value resets it to zero.
	if kind == reflect.Int64 && item.Type() == durationType {
		if value == "" {
			item.SetInt(0)
			return nil
		}

		d, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("'%v' is incorrect duration", value)
		}

		item.SetInt(int64(d))
		return nil
	}

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
		// Pointer to the url.URL struct only (other types are rejected
		// earlier by getFieldCastList).
		u, err := url.Parse(value)
		if err != nil {
			return err
		}

		item.Set(reflect.ValueOf(u))
	case reflect.Struct:
		// The url.URL struct only (other types are rejected earlier
		// by getFieldCastList).
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
			return 0, fmt.Errorf("%d overflows uint32", r)
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
		// ±Inf is representable in float32, so it is not an overflow; only a
		// finite magnitude beyond MaxFloat32 is.
		if !math.IsInf(r, 0) && math.Abs(r) > math.MaxFloat32 {
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

// strToBool converts string to bool type. An empty string yields false.
// Only the literals accepted by strconv.ParseBool are valid: 1, t, T, TRUE,
// true, True, 0, f, F, FALSE, false, False.
func strToBool(value string) (bool, error) {
	// For empty string returns false.
	if len(value) == 0 {
		return false, nil
	}

	r, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf(
			"'%v' has incorrect type, bool expected",
			value,
		)
	}

	return r, nil
}
