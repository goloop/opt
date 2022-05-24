package opt

import (
	"fmt"
	"math"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// The testUIFDataTestType the uint, int and float test type.
type testUIFDataTestType struct {
	Value   string
	Control string
	Correct bool
	Kind    reflect.Kind
}

// The testBoolDataTestType the boolean test type.
type testBoolDataTestType struct {
	Value   string
	Control bool
	Correct bool
}

// TestUnmarshalOpt is classic test for unmarshalOpt function.
func TestUnmarshalOpt(t *testing.T) {
	type data struct {
		Help  bool   `opt:"h" alt:"help" help:"show help information"`
		Debug bool   `opt:"debug" alt:"d" def:"true" help:"debug mode"`
		Host  string `opt:"host" def:"localhost" help:"host name"`
		Port  int    `opt:"p" alt:"port" def:"8080" help:"port number"`

		MemorySize int    `def:"1024"` // default value
		UserName   string // default long flag user-name
		Ignored    bool   `opt:"-"` // ignored
	}

	// Data for testing.
	exp := "false:false:0.0.0.0:8000:Goloop:1024"
	tests := [][]string{
		split("/app:--host=0.0.0.0:-p8000:-d:false:--user-name:Goloop"),
		split("/app:--host:0.0.0.0:--user-name:Goloop:-p:8000:-dfalse"),
		split("/app:--user-name:Goloop:-p:8000:--host:0.0.0.0:--no-debug"),
		split("/app:-p:8000:--host:0.0.0.0:--user-name:Goloop:-d:false"),
	}

	// Testing.
	for i, args := range tests {
		d := data{}
		if err := unmarshalOpt(&d, args); err != nil {
			t.Error(err)
		}

		v := fmt.Sprintf(
			"%v:%v:%s:%d:%s:%d",
			d.Help, d.Debug, d.Host, d.Port, d.UserName, d.MemorySize,
		)
		if v != exp {
			t.Errorf("%d test, expected %v but %v", i, exp, v)
		}
	}
}

// TestUnmarshalOptMix tests unmarshalOpt with mixed flags.
// Changes the order of the flags, uppercase and lowercase for long flags etc.
func TestUnmarshalOptMix(t *testing.T) {
	type data struct {
		Verbose  bool     `opt:"v" alt:"verbose" help:"verbose output"`
		Debug    bool     `opt:"debug" alt:"d" help:"debug mode"`
		Users    []string `opt:"users" alt:"U" def:"John,Bob,Robert" sep:","`
		Geometry string   `opt:"g" def:"T B"`
		Path     string   `opt:"0"`
		A        int      `opt:"1" def:"5"`
		B        int      `opt:"2" def:"10"`
		C        int      `opt:"3" def:"15"`
	}

	tests := [][]string{
		split("./app:-dUJack:-U:Bob:--users=Roy:--no-verbose:-gL R:5:10:15"),
		split("./app:-dU:Jack:-U:Bob:--USERS=Roy:--No-Verbose:-g:L R:5:10:15"),
		split("./app:-d:--users:Jack,Bob,Roy:--no-verbose:-gL R:5:10:15"),
		split("./app:-g:L R:-U:Jack:-UBob:-URoy:--no-verbose:-d:--:5:10:15"),
		split("./app:5:-dU:Jack:--users:Bob,Roy:--verbose:false:-gL R:10:15"),
		split("./app:5:10:--no-verboSe:-dU:Jack:-U:Bob:-U:Roy:-g:L R:15"),
		split("./app:5:10:15:-UJack:-UBob:-URoy:--verbose=false:-g:L R:-d"),
		split("./app:5:10:15:-U:Jack:-UBob:-URoy:-d:--VERBOSE:false:-gL R"),
		split("./app:5:-UJack:--no-verbose:-gL R:-UBob:-d:true:-URoy:10:15"),
		split("./app:5:-UJack:-UBob:--verbose:false:-URoy:-gL R:-d:--:10:15"),
		split("./app:5:-UJack,Bob,Roy:-gL R:-d:--no-verbose:--:10:15"),
		split("./app:5:-UJack:--no-verbose:-gL R:-UBob:-d:true:-URoy:10"),
		split("./app:-UJack:-UBob:--verbose:false:-URoy:-gL R:-d:--:5"),
		split("./app:-UJack,Bob,Roy:-gL R:-d:--no-verbose"),
	}

	exp := data{
		Verbose:  false,
		Debug:    true,
		Users:    []string{"Jack", "Bob", "Roy"},
		Geometry: "L R",
		Path:     "./app",
		A:        5,
		B:        10,
		C:        15,
	}

	for i, test := range tests {
		var d = data{}
		if err := unmarshalOpt(&d, test); err != nil {
			t.Error(err)
			return // it makes no sense to display all errors
		}

		if !reflect.DeepEqual(exp, d) {
			t.Errorf("%d test, expected %v but %v", i, exp, d)
		}
	}
}

// TestWrongArgs tests with wrong value.
func TestWrongArgs(t *testing.T) {
	type data struct {
		Words   []string `opt:"w" alt:"word" def:"One Two Three Four"`
		Debug   bool     `opt:"debug" alt:"d" help:"debug mode"`
		InSlice []int    `opt:"s" def:"1,2,3" sep:","`
		InArray [3]int   `opt:"a" def:"1,2,3" sep:","`
		Uint    uint     `def:"1"`
		Float   float64  `def:"1.3"`
	}

	split := func(str string) []string { return strings.Split(str, ":") }

	// Undeclared tag.
	obj := data{}
	test := split("./app:-d:--user:John") // -U isn't declared
	if err := unmarshalOpt(&obj, test); err == nil {
		t.Error("expected error")
	}

	// List to single value.
	obj = data{}
	test = split("./app:-d:true:-d:-dfalse") // -d should be false
	if err := unmarshalOpt(&obj, test); err != nil {
		t.Error(err)
	}

	if obj.Debug {
		t.Error("expected true")
	}

	// Slice and Array.
	obj = data{}
	test = split("./app:-s:7:-s:5:-s3:-a:7,5,3")
	if err := unmarshalOpt(&obj, test); err != nil {
		t.Error(err)
	}

	if e := []int{7, 5, 3}; !reflect.DeepEqual(e, obj.InSlice) {
		t.Errorf("expected %v but %v", e, obj.InSlice)
	}

	if e := [3]int{7, 5, 3}; !reflect.DeepEqual(e, obj.InArray) {
		t.Errorf("expected %v but %v", e, obj.InSlice)
	}

	// Default Slice and Array.
	obj = data{}
	test = split("./app:-a:7,5:-a3")
	if err := unmarshalOpt(&obj, test); err != nil {
		t.Error(err)
	}

	if e := []int{1, 2, 3}; !reflect.DeepEqual(e, obj.InSlice) {
		t.Errorf("expected %v but %v", e, obj.InSlice)
	}

	if e := [3]int{7, 5, 3}; !reflect.DeepEqual(e, obj.InArray) {
		t.Errorf("expected %v but %v", e, obj.InSlice)
	}

	if obj.Uint != 1 {
		t.Errorf("expected %v but %v", 1, obj.Uint)
	}

	if obj.Float != 1.3 {
		t.Errorf("expected %v but %v", 1.3, obj.Float)
	}

	// Array overflow..
	obj = data{}
	test = split("./app:-a:7,5:-a3,1")
	if err := unmarshalOpt(&obj, test); err == nil {
		t.Error("expected error")
	}

	// Slice and Array incorrect values.
	obj = data{}
	test = split("./app:-s:a,b,c")
	if err := unmarshalOpt(&obj, test); err == nil {
		t.Error("expected error")
	}

	test = split("./app:-a:a,b:-ac")
	if err := unmarshalOpt(&obj, test); err == nil {
		t.Error("expected error")
	}
}

// TestUnmarshalOptLongName tests unmarshalOpt with a long name.
func TestUnmarshalOptLongName(t *testing.T) {
	// Note: An incorrect storage object causes the function to cause panic!
	defer func() {
		if err := recover(); err == nil {
			t.Error("an error is expected for long option name")
		}
	}()

	type data struct { // |******************************| ______ max 32 chars
		Name string `opt:"very-lon-flag-name-one-two-three-x"`
	}
	unmarshalOpt(&data{}, []string{"/app", "5", "10"}) // panic is expected
}

// TestUnmarshalOptNotPointer tests unmarshalOpt with a non-pointer object.
func TestUnmarshalOptNotPointer(t *testing.T) {
	// Note: An incorrect storage object causes the function to cause panic!
	defer func() {
		if err := recover(); err == nil {
			t.Error("an error is expected for non-pointer object")
		}
	}()

	type data struct{}
	unmarshalOpt(data{}, []string{"/app", "5", "10"}) // panic is expected
}

// TestUnmarshalOptNotInitialized tests unmarshalOpt
// with a not initialized object.
func TestUnmarshalOptNotInitialized(t *testing.T) {
	// Note: An incorrect storage object causes the function to cause panic!
	defer func() {
		if err := recover(); err == nil {
			t.Error("an error is expected for not initialized object")
		}
	}()

	var d *struct{}
	unmarshalOpt(d, []string{"/app", "5", "10", "15"}) // panic is expected
}

// TestUnmarshalOptNotStruct tests unmarshalOpt
// with a object that isn't a struct.
func TestUnmarshalOptNotStruct(t *testing.T) {
	// Note: An incorrect storage object causes the function to cause panic!
	defer func() {
		if err := recover(); err == nil {
			t.Error("an error is expected for a non-struct pointer")
		}
	}()

	var d = new(int)
	unmarshalOpt(d, []string{"/app", "5", "10"}) // panic is expected
}

// TestOverflow tests with struct.
func TestOverflow(t *testing.T) {
	split := func(str string) []string { return strings.Split(str, ":") }

	test := split("./app:--int:999999999999999999999999999999999999999999999")
	i := struct{ Int int }{}
	if err := unmarshalOpt(&i, test); err == nil {
		t.Error("expected an error")
	}

	test = split("./app:--uint:999999999999999999999999999999999999999999999")
	u := struct{ Uint uint }{}
	if err := unmarshalOpt(&u, test); err == nil {
		t.Error("expected an error")
	}

	test = split("./app:--float:99999999999999999999999999999999999999999999")
	f := struct{ Float float32 }{}
	if err := unmarshalOpt(&f, test); err == nil {
		t.Error("expected an error")
	}
}

// TestUnmarshalOptBoolWrong tests unmarshalOpt with wrong bool.
func TestUnmarshalOptBoolWrong(t *testing.T) {
	var b bool
	split := func(str string) []string { return strings.Split(str, ":") }
	objPtr := struct {
		Bool *bool
	}{&b}

	test := split("./app:--bool:yes")
	if err := unmarshalOpt(&objPtr, test); err == nil {
		t.Error("expected error")
	}

	// ...
	objElem := struct {
		Bool bool
	}{}

	test = split("./app:--bool:yes")
	if err := unmarshalOpt(&objElem, test); err == nil {
		t.Error("expected error")
	}
}

// TestWrongList tests with wrong list.
func TestWrongList(t *testing.T) {
	split := func(str string) []string { return strings.Split(str, ":") }
	objSlice := struct {
		Slice []chan int `opt:"slice" def:"a"`
	}{}

	test := split("./app:--slice:c")
	if err := unmarshalOpt(&objSlice, test); err == nil {
		t.Error("expected error")
	}

	objArray := struct {
		Array []chan int `opt:"array" def:"a"`
	}{}

	test = split("./app:--array:c")
	if err := unmarshalOpt(&objArray, test); err == nil {
		t.Error("expected error")
	}
}

// TestStructUrl tests with pointer on struct.
func TestStructUrl(t *testing.T) {
	var (
		site url.URL
	)

	split := func(str string) []string { return strings.Split(str, ":") }
	test := split("./app:--site:%$")

	// Pointer on the url.URL.
	objOk := struct {
		Site *url.URL `opt:"site"`
	}{
		Site: &site,
	}
	if err := unmarshalOpt(&objOk, test); err == nil {
		t.Error("expected en error")
	}

	objElem := struct {
		Site url.URL `opt:"site"`
	}{}
	if err := unmarshalOpt(&objElem, test); err == nil {
		t.Error("expected en error")
	}
}

// TestStruct tests with struct.
func TestStruct(t *testing.T) {
	type data struct{}

	split := func(str string) []string { return strings.Split(str, ":") }
	test := split("./app:--site:example.com:--site:example.com")

	// Pointer on the url.URL.
	objOk := struct {
		Site url.URL `opt:"site"`
	}{}
	if err := unmarshalOpt(&objOk, test); err != nil {
		t.Error(err)
	}

	// Panic for struct.
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected error for pointer ont the struct")
		}
	}()

	objErr := struct {
		Any data `opt:"any"`
	}{}
	unmarshalOpt(&objErr, test)
}

// TestPtr tests with pointers.
func TestPtr(t *testing.T) {
	type data struct {
		Debug bool `opt:"debug" alt:"d" help:"debug mode"`
		Age   *int `opt:"age"`
	}

	var age int
	split := func(str string) []string { return strings.Split(str, ":") }

	// Pointer int.
	obj := data{Age: &age}
	test := split("./app:--age:99")
	if err := unmarshalOpt(&obj, test); err != nil {
		t.Error(err)
	}

	// Not initialized pointer.
	obj = data{}
	test = split("./app:--age:99")
	if err := unmarshalOpt(&obj, test); err == nil {
		t.Error("expected error")
	}
}

// TestStructPtr tests with pointer on struct.
func TestStructPtr(t *testing.T) {
	type data struct{}

	var (
		site url.URL
		any  data

		i int
		u uint
		f float32
	)

	split := func(str string) []string { return strings.Split(str, ":") }
	test := split("./app:--site:example.com:--site:example.com:" +
		"--uint:3:--int:5:--float:7")

	// Pointer on the url.URL.
	objOk := struct {
		Site  *url.URL `opt:"site"`
		Uint  *uint
		Int   *int
		Float *float32
	}{
		Site:  &site,
		Int:   &i,
		Uint:  &u,
		Float: &f,
	}
	if err := unmarshalOpt(&objOk, test); err != nil {
		t.Error(err)
	}

	test = split("./app:--site:10")
	if err := unmarshalOpt(&objOk, test); err != nil {
		t.Error(err)
	}

	// Panic for struct.
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected error for pointer on the struct")
		}
	}()

	objErr := struct {
		Any *data `opt:"any"`
	}{
		Any: &any,
	}
	unmarshalOpt(&objErr, test)
}

// TestUnmarshalOptPositional tests unmarshalOpt for positional arguments.
func TestUnmarshalOptPositional(t *testing.T) {
	type data struct {
		Path string `opt:"0" help:"path to bin file"`
		A    int    `opt:"1"`
		B    int    `opt:"2"`
		C    int    `opt:"3"`
		ABC  []int  `opt:"[]" help:"position args as list"`
	}

	var d, sum = data{}, 0

	if err := unmarshalOpt(&d, []string{"/app", "5", "10", "15"}); err != nil {
		t.Error(err)
	}

	// Calculate sum of positional arguments.
	for _, i := range d.ABC {
		sum += i
	}

	// Check the correctness of the data in single cells
	// and in the list of arguments.
	if d.A+d.B+d.C != sum || sum == 0 {
		t.Error("incorrect parser by positional argument:", d.ABC)
	}

	// Check path value.
	if d.Path != "/app" {
		t.Error("incorrect parser bin path:", d.Path)
	}

	// Checking the sequence of arguments.
	if d.A != 5 || d.B != 10 {
		t.Error("order expected 5 10 15 but", d.A, d.B, d.C)
	}

	if tmp := []int{5, 10, 15}; !reflect.DeepEqual(tmp, d.ABC) {
		t.Errorf("order expected %v but %v", tmp, d.ABC)
	}
}

// TestStrToBool tests strToBool function.
func TestStrToBool(t *testing.T) {
	var tests = []testBoolDataTestType{
		{"", false, true},
		{"0", false, true},
		{"1", true, true},
		{"1.1", true, true},
		{"-1.1", true, true},
		{"0.0", false, true},
		{"true", true, true},
		{"True", true, true},
		{"TRUE", true, true},
		{"false", false, true},
		{"False", false, true},
		{"FALSE", false, true},
		{"string", false, false},
		{"a:b:c", false, false},
	}

	// Test correct values.
	for _, test := range tests {
		r, err := strToBool(test.Value)
		if test.Correct && err != nil {
			t.Error(err)
		} else if !test.Correct && err == nil {
			t.Errorf("value %s should throw an exception", test.Value)
		}

		if r != test.Control {
			t.Errorf("expected %s but the result %t", test.Value, r)
		}
	}
}

// TestStrToIntKind tests strToIntKind function.
func TestStrToIntKind(t *testing.T) {
	var (
		tests    []testUIFDataTestType
		maxInt   = fmt.Sprintf("%d", math.MaxInt64-1)
		maxInt8  = fmt.Sprintf("%d", math.MaxInt8-1)
		maxInt16 = fmt.Sprintf("%d", math.MaxInt16-1)
		maxInt32 = fmt.Sprintf("%d", math.MaxInt32-1)
		maxInt64 = fmt.Sprintf("%d", math.MaxInt64-1)
	)

	// For 32-bit platform.
	if strconv.IntSize == 32 {
		maxInt = maxInt32
	}

	// Test data.
	tests = []testUIFDataTestType{
		{"", "0", true, reflect.Int},
		{"0", "0", true, reflect.Int},
		{"-3", "-3", true, reflect.Int},
		{"3", "3", true, reflect.Int},

		{"-128", "-128", true, reflect.Int8},
		{"127", "127", true, reflect.Int8},

		{maxInt, maxInt, true, reflect.Int},
		{maxInt8, maxInt8, true, reflect.Int8},
		{maxInt16, maxInt16, true, reflect.Int16},
		{maxInt32, maxInt32, true, reflect.Int32},
		{maxInt64, maxInt64, true, reflect.Int64},

		{"string", "0", false, reflect.Int},
		{"3" + maxInt, "0", false, reflect.Int},
		{"3" + maxInt8, "0", false, reflect.Int8},
		{"-129", "0", false, reflect.Int8},
		{"128", "0", false, reflect.Int8},
		{maxInt16 + "0", "0", false, reflect.Int16},
		{maxInt32 + "0", "0", false, reflect.Int32},
		{maxInt64 + "0", "0", false, reflect.Int64},
		{"0", "0", false, reflect.Slice},
	}

	// Test correct values.
	for _, data := range tests {
		r, err := strToIntKind(data.Value, data.Kind)
		if data.Correct && err != nil {
			t.Error(err)
		} else if !data.Correct && err == nil {
			t.Errorf("the value %s should throw an exception", data.Value)
		} else if err != nil && r != 0 {
			t.Errorf("any error should return zero but returns %v", r)
		}

		control := fmt.Sprintf("%d", r)
		if control != data.Control {
			t.Errorf("expected %s but returns %s", data.Control, control)
		}
	}
}

// TestStrToUintKind tests strToUintKind function.
func TestStrToUintKind(t *testing.T) {
	var (
		tests     []testUIFDataTestType
		maxUint   = "18446744073709551614"
		maxUint8  = fmt.Sprintf("%d", math.MaxUint8-1)
		maxUint16 = fmt.Sprintf("%d", math.MaxUint16-1)
		maxUint32 = fmt.Sprintf("%d", math.MaxUint32-1)
		maxUint64 = "18446744073709551614"
	)

	// For 32-bit platform.
	if strconv.IntSize == 32 {
		maxUint = maxUint32
	}

	// Test data.
	tests = []testUIFDataTestType{
		{"", "0", true, reflect.Uint},
		{"0", "0", true, reflect.Uint},
		{"3", "3", true, reflect.Uint},
		{maxUint, maxUint, true, reflect.Uint},
		{maxUint8, maxUint8, true, reflect.Uint8},
		{maxUint16, maxUint16, true, reflect.Uint16},
		{maxUint32, maxUint32, true, reflect.Uint32},
		{maxUint64, maxUint64, true, reflect.Uint64},

		{"string", "0", false, reflect.Uint},
		{"-3", "0", false, reflect.Uint},
		{"9" + maxUint, "0", false, reflect.Uint},
		{"9" + maxUint8, "0", false, reflect.Uint8},
		{"9" + maxUint16, "0", false, reflect.Uint16},
		{"9" + maxUint32, "0", false, reflect.Uint32},
		{"9" + maxUint64, "0", false, reflect.Uint64},
		{"0", "0", false, reflect.Slice},
	}

	// Test correct values.
	for _, data := range tests {
		r, err := strToUintKind(data.Value, data.Kind)
		if data.Correct && err != nil {
			t.Error(err)
		} else if !data.Correct && err == nil {
			t.Errorf("the value %s should throw an exception", data.Value)
		} else if err != nil && r != 0 {
			t.Errorf("any error should return zero but returns %v", r)
		}

		control := fmt.Sprintf("%d", r)
		if control != data.Control {
			t.Errorf("expected %s but generated %s", data.Control, control)
		}
	}
}

// TestStrToFloatKind tests strToFloatKind function.
func TestStrToFloatKind(t *testing.T) {
	var (
		tests      []testUIFDataTestType
		maxFloat32 = fmt.Sprintf("%.2f", math.MaxFloat32-1)
		maxFloat64 = fmt.Sprintf("%.2f", math.MaxFloat64-1)
	)

	// Test data.
	tests = []testUIFDataTestType{
		{"", "0.00", true, reflect.Float64},
		{"0.0", "0.00", true, reflect.Float64},
		{"3.0", "3.00", true, reflect.Float64},
		{"-3.1", "-3.10", true, reflect.Float64},
		{maxFloat32, maxFloat32, true, reflect.Float32},
		{maxFloat64, maxFloat64, true, reflect.Float64},

		{"string", "0.00", false, reflect.Float64},
		{"9" + maxFloat32, "0.00", false, reflect.Float32},
		{"9" + maxFloat64, "0.00", false, reflect.Float64},
		{"0.00", "0.00", false, reflect.Slice},
	}

	// Test correct values.
	for _, data := range tests {
		r, err := strToFloatKind(data.Value, data.Kind)
		if data.Correct && err != nil {
			t.Error(err)
		} else if !data.Correct && err == nil {
			t.Errorf("the value %s should throw an exception", data.Value)
		} else if err != nil && r != 0 {
			t.Errorf("any error should return zero but returns %v", r)
		}

		control := fmt.Sprintf("%.2f", r)
		if control != data.Control {
			t.Errorf("expected %s but generated %s", data.Control, control)
		}
	}
}
