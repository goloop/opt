package opt

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// unmarshalOPT gets variables from the argument-line and sets them
// into object by pointer. Returns an error if something went wrong.
//
// unmarshalOPT method supports the following field's types: int, int8, int16,
// int32, int64, uin, uint8, uin16, uint32, in64, float32, float64, string,
// bool, url.URL and pointers, array or slice from thous types (i.e. *int, ...,
// []int, ..., []bool, ..., [2]*url.URL, etc.).
//
// For other filed's types (like chan, map ...) will be returned an error.
func unmarshalOPT(obj interface{}, args []string) error {
	var rt, rv = reflect.TypeOf(obj), reflect.ValueOf(obj)

	// The obj argument should be a pointer to initialized object.
	if obj == nil ||
		rt.Kind() != reflect.Ptr || // check for pointer first ...
		rt.Elem().Kind() != reflect.Struct || // ... and after on the struct
		!rv.Elem().IsValid() {
		return errors.New("cannot unmarshal command-line arguments " +
			"into not *struct type")
	}

	// Walk through all the fields of the structure and read all tags.
	sops := ""
	fields := []*fieldSample{}
	elem := rv.Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := rt.Elem().Field(i)
		ts := getTagSample(field.Tag.Get("opt"), field.Name)
		item := elem.FieldByName(field.Name)
		fields = append(fields, &fieldSample{ts, &item})

		if ts.Short != "" {
			sops += ts.Short
		}
	}

	// Parse options.
	opts, err := getOptions(args, sops)
	if err != nil {
		return err
	}

	for _, f := range fields {
		if f.TagSample.Short == "?" {
			// Generate help info.
			help := getHelp(rv, fields, opts)
			f.Item.Set(reflect.ValueOf(help))
			continue
		}

		// Update value.
		err := updateByTag(f, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

// The updateByTag updates the data for
// the field (access to the field via a tag).
func updateByTag(field *fieldSample, opts optSamples) error {
	var (
		sep   = ","
		seq   = []string{}
		value = field.TagSample.Value
		kind  = field.Item.Kind()
	)

	kindIsString := kind == reflect.String
	kindIsSequence := kind == reflect.Array || kind == reflect.Slice
	if field.TagSample.Short == "?" && !kindIsString {
		return errors.New("container for help data must be string type")
	} else if field.TagSample.Short == "[]" && !kindIsSequence {
		return fmt.Errorf("the container for positional arguments "+
			"must be a slice or an array but not %v", kind)
	}

	if field.TagSample.Short == "[]" {
		value = ""
		seq = opts.PositionalValues()
	} else if v, ok := opts[field.TagSample.Short]; ok {
		value = v // by short option name
	} else if v, ok := opts[field.TagSample.Long]; ok {
		value = v // by long option name
	}

	// Set values of the desired type.
	switch kind {
	case reflect.Array:
		if len(seq) == 0 {
			seq = strings.Split(value, sep)
		}

		if max := field.Item.Type().Len(); len(seq) > max {
			kind := field.Item.Index(0).Kind()
			return fmt.Errorf("%d items overflow [%d]%v array",
				len(seq), max, kind)
		}

		err := setSequence(field.Item, seq)
		if err != nil {
			return err
		}
	case reflect.Slice:
		if len(seq) == 0 {
			seq = strings.Split(value, sep)
		}

		tmp := reflect.MakeSlice(field.Item.Type(), len(seq), len(seq))
		err := setSequence(&tmp, seq)
		if err != nil {
			return err
		}
		field.Item.Set(reflect.AppendSlice(*field.Item, tmp))
	case reflect.Ptr:
		k, t := field.Item.Type().Elem().Kind(), field.Item.Type()
		if k == reflect.Struct && t != reflect.TypeOf((*url.URL)(nil)) {
			return fmt.Errorf("%s field has invalid type",
				field.Item.Type().Name())
		}

		if k != reflect.Struct {
			// If the pointer is not to a structure.
			tmp := reflect.Indirect(*field.Item)
			err := setValue(tmp, value)
			if err != nil {
				return err
			}
		} else {
			// If a pointer to a structure of the url.URL.
			err := setValue(*field.Item, value)
			if err != nil {
				return err
			}
		}
	case reflect.Struct:
		if field.Item.Type() != reflect.TypeOf(url.URL{}) {
			return fmt.Errorf("%s field has invalid type",
				field.Item.Type().Name())
		}

		err := setValue(*field.Item, value) // if a url.URL structure
		if err != nil {
			return err
		}
	default:
		err := setValue(*field.Item, value)
		if err != nil {
			return err
		}
	}

	return nil
}

// The setSequence sets slice into item.
func setSequence(item *reflect.Value, seq []string) (err error) {
	var kind = item.Index(0).Kind()

	defer func() {
		// Catch the panic and return an exception as a value.
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	// Ignore empty containers.
	switch {
	case kind == reflect.Array && item.Type().Len() == 0:
		fallthrough
	case kind == reflect.Slice && item.Len() == 0:
		fallthrough
	case len(seq) == 0:
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

// The setValue sets value into item.
func setValue(item reflect.Value, value string) error {
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
		if item.Type() != reflect.TypeOf((*url.URL)(nil)) {
			return fmt.Errorf("%s field has invalid type", item.Type().Name())
		}

		// The url.URL struct only.
		u, err := url.Parse(value)
		if err != nil {
			return err
		}

		item.Set(reflect.ValueOf(u))
	case reflect.Struct:
		if item.Type() != reflect.TypeOf(url.URL{}) {
			return fmt.Errorf("%s field has invalid type", item.Type().Name())
		}

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

// strToIntKind convert string to int64 type with checking for conversion
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
		return 0, err
	}

	switch kind {
	case reflect.Int:
		// For 32-bit platform it is necessary to check overflow.
		if strconv.IntSize == 32 {
			if r < math.MinInt32 || r > math.MaxInt32 {
				return 0, fmt.Errorf("%d overflows int (int32)", r)
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
		// pass
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
		return 0, err
	}

	switch kind {
	case reflect.Uint:
		// For 32-bit platform it is necessary to check overflow.
		if strconv.IntSize == 32 && r > math.MaxUint32 {
			return 0, fmt.Errorf("%d overflows uint (uint32)", r)
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
		// pass
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
		return 0.0, err
	}

	switch kind {
	case reflect.Float32:
		if math.Abs(r) > math.MaxFloat32 {
			return 0.0, fmt.Errorf("%f overflows float32", r)
		}
	case reflect.Float64:
		// pass
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
			return r, errB
		}

		if math.Abs(f) > epsilon {
			r = true
		}
	}

	return r, nil
}
