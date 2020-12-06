package opt

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"
)

// UIFDataTestType the uint, int and float test type.
type UIFDataTestType struct {
	Value   string
	Control string
	Correct bool
	Kind    reflect.Kind
}

// BoolDataTestType the boolean test type.
type BoolDataTestType struct {
	Value   string
	Control bool
	Correct bool
}

// TestUnmarshalOPT classic test for unmarshalOPT function.
func TestUnmarshalOPT(t *testing.T) {
	type data struct {
		Help  bool   `opt:"h,help,,show help information"`
		Debug bool   `opt:"d,debug,true,debug mode"`
		Host  string `opt:",host,localhost,host name"`
		Port  int    `opt:"p,port,8080,port number"`
	}

	// Data for testing.
	var tests = [][]string{
		arg("/main:--host=0.0.0.0:-p8000:-d:false"),
		arg("/main:--host:0.0.0.0:-p:8000:-dfalse"),
		arg("/main:-p:8000:--host:0.0.0.0:--no-debug"),
		arg("/main:-p:8000:--host:0.0.0.0:-d:false"),
	}

	// Testing.
	for _, args := range tests {
		d := data{}
		err := unmarshalOPT(&d, args)
		if err != nil {
			t.Error(err)
		}

		v := fmt.Sprintf("h:%t;d:%t;h:%s;p:%d", d.Help,
			d.Debug, d.Host, d.Port)
		if v != "h:false;d:false;h:0.0.0.0;p:8000" {
			t.Error("incorrect:", v)
		}
	}
}

// TestUnmarshalOPTNotPointer tests unmarshalOPT for the correct handling
// of an exception for a non-pointer value.
func TestUnmarshalOPTNotPointer(t *testing.T) {
	type data struct{}
	if err := unmarshalOPT(data{}, []string{"/main", "5", "10"}); err == nil {
		t.Error("an error is expected for non-pointer value")
	}
}

// TestUnmarshalOPTNotInitialized tests unmarshalOPT for the correct handling
// of an exception for a not initialized value.
func TestUnmarshalOPTNotInitialized(t *testing.T) {
	type data struct{}
	var d *data
	if err := unmarshalOPT(d, []string{"/main", "5", "10", "15"}); err == nil {
		t.Error("an error is expected for not initialized value")
	}
}

// TestUnmarshalOPTNotStruct tests unmarshalOPT for the correct handling
// of an exception for a value that isn't a struct.
func TestUnmarshalOPTNotStruct(t *testing.T) {
	var d = new(int)
	if err := unmarshalOPT(d, []string{"/main", "5", "10"}); err == nil {
		t.Error("an error is expected for a pointer not to a struct")
	}
}

// TestUnmarshalOPTPositional tests unmarshalOPT for positional arguments.
func TestUnmarshalOPTPositional(t *testing.T) {
	type data struct {
		Path string `opt:"0,,,path to bin file"`
		A    int    `opt:"1"`
		B    int    `opt:"2"`
		C    int    `opt:"3"`
		ABC  []int  `opt:"[],,,position args as list"`
	}

	var d, sum = data{}, 0

	err := unmarshalOPT(&d, []string{"/main", "5", "10", "15"})
	if err != nil {
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
	if d.Path != "/main" {
		t.Error("incorrect parser bin path:", d.Path)
	}

	// Checking the sequence of arguments.
	if d.A != 5 || d.B != 10 {
		t.Error("order expected 5 10 15 but", d.A, d.B, d.C)
	}

	if sts(d.ABC, " ") != "5 10 15" {
		t.Error("order expected 5 10 15 but", sts(d.ABC, " "))
	}
}

// TestUnmarshalOPTMix random opt testing.
func TestUnmarshalOPTMix(t *testing.T) {
	type data struct {
		Verbose  bool     `opt:"v,verb,true,verbose output"`
		Debug    bool     `opt:"d,debug,,debug mode"`
		Users    []string `opt:"U,users,{John,Bob,Robert},user list"`
		Greeting string   `opt:"g,,'Hello, world!',greeting message"`
		Path     string   `opt:"0,,,path to bin"`
		A        int      `opt:"1"`
		B        int      `opt:"2"`
		C        int      `opt:"3"`
		ABC      []int    `opt:"[]"`
	}

	// Data for testing.
	var tests = [][]string{
		arg("/main:-dU Jack,Harry:--no-verb:-g Hi, all:5:10:15"),
		arg("/main:-d:-U Jack,Harry:--no-verb:-g Hi, all:--:5:10:15"),
		arg("/main:-dU Jack,Harry:--verb false:-g Hi, all:5:10:15"),
		arg("/main:-dU Jack,Harry:--no-verb true:-g Hi, all:5:10:15"),
		arg("/main:-dUJack,Harry:--verb=false:-gHi, all:5:10:15"),
		arg("/main:5:10:15:-dUJack,Harry:--verb false:-gHi, all"),
		arg("/main:5:10:15:-UJack,Harry:--verb false:-gHi, all:-d true"),
		arg("/main:5:-UJack,Harry:--verb false:-gHi, all:-d:--:10:15"),
	}

	// Correct tests.
	for _, test := range tests {
		var d = data{}
		err := unmarshalOPT(&d, test)
		if err != nil {
			t.Error(err)
		}

		// Check fields:
		if d.Verbose {
			t.Errorf("incorrect d.Verbose==%t for: %v", d.Verbose, test)
		}

		if !d.Debug {
			t.Errorf("incorrect d.Debug==%t for: %v", d.Debug, test)
		}

		if sts(d.Users, ":") != sts([]string{"Jack", "Harry"}, ":") {
			t.Errorf("incorrect d.Users==%v for: %v", d.Users, test)
		}

		if d.Greeting != "Hi, all" {
			t.Errorf("incorrect d.Greeting==%s for: %v", d.Greeting, test)
		}

		if (d.A != 5 || d.B != 10) || d.C != 15 {
			t.Errorf("incorrect d.A==%d d.B==%d d.C==%d for: %v",
				d.A, d.B, d.C, test)
		}

		if sts(d.ABC, ":") != sts([]int{5, 10, 15}, ":") {
			t.Errorf("incorrect d.ABC==%v for: %v", d.ABC, test)
		}

		if d.Path != "/main" {
			t.Error("incorrect parser bin path:", d.Path)
		}
	}
}

// TestStrToBool tests strToBool function.
func TestStrToBool(t *testing.T) {
	var tests = []BoolDataTestType{
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
		tests    []UIFDataTestType
		maxInt   string = fmt.Sprintf("%d", math.MaxInt64-1)
		maxInt8  string = fmt.Sprintf("%d", math.MaxInt8-1)
		maxInt16 string = fmt.Sprintf("%d", math.MaxInt16-1)
		maxInt32 string = fmt.Sprintf("%d", math.MaxInt32-1)
		maxInt64 string = fmt.Sprintf("%d", math.MaxInt64-1)
	)

	// For 32-bit platform.
	if strconv.IntSize == 32 {
		maxInt = maxInt32
	}

	// Test data.
	tests = []UIFDataTestType{
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
		{"3" + maxInt16, "0", false, reflect.Int16},
		{"3" + maxInt32, "0", false, reflect.Int32},
		{"3" + maxInt64, "0", false, reflect.Int64},
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
		tests     []UIFDataTestType
		maxUint   string = "18446744073709551614"
		maxUint8  string = fmt.Sprintf("%d", math.MaxUint8-1)
		maxUint16 string = fmt.Sprintf("%d", math.MaxUint16-1)
		maxUint32 string = fmt.Sprintf("%d", math.MaxUint32-1)
		maxUint64 string = "18446744073709551614"
	)

	// For 32-bit platform.
	if strconv.IntSize == 32 {
		maxUint = maxUint32
	}

	// Test data.
	tests = []UIFDataTestType{
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
		tests      []UIFDataTestType
		maxFloat32 string = fmt.Sprintf("%.2f", math.MaxFloat32-1)
		maxFloat64 string = fmt.Sprintf("%.2f", math.MaxFloat64-1)
	)

	// Test data.
	tests = []UIFDataTestType{
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
