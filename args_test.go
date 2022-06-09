package opt

import (
	"reflect"
	"strings"
	"testing"
)

// TestParse tests parse method.
func TestParse(t *testing.T) {
	split := func(str string) []string { return strings.Split(str, ":") }
	flags := map[string]int{"d": 1, "U": 1, "g": 1, "verbose": 1}
	expected := map[string][]string{
		"0": {"./app"}, "1": {"5"}, "2": {"10"}, "3": {"15"},
		"U": {"Jan,Bob"}, "d": {"true"}, "g": {"Hello, world"},
		"verbose": {"false"},
	}
	tests := [][]string{
		split("./app:-dU:Jan,Bob:--no-verbose:-g:Hello, world:5:10:15"),
		split("./app:-d:-U:Jan,Bob:--no-verbose:-g:Hello, world:--:5:10:15"),
		split("./app:-dU:Jan,Bob:--verbose:false:-g:Hello, world:5:10:15"),
		split("./app:-dUJan,Bob:--verbose=false:-gHello, world:5:10:15"),
		split("./app:5:10:15:-dUJan,Bob:--verbose:false:-gHello, world"),
		split("./app:5:10:15:-UJan,Bob:--verbose:false:-gHello, world:-d:true"),
		split("./app:5:-UJan,Bob:--verbose:false:-gHello, world:-d:--:10:15"),
	}

	am := argMap{}
	for _, test := range tests {
		if err := am.parse(test, flags); err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(am.asFlat(), expected) {
			t.Errorf("expected %v but %v", expected, am)
		}
	}
}

// TestParseMultiflags tests parse method with multiflags.
func TestParseMultiflags(t *testing.T) {
	split := func(str string) []string { return strings.Split(str, ":") }
	flags := map[string]int{"U": 1, "verbose": 1}
	expected := map[string][]string{
		"0": {"./app"}, "1": {"5"}, "2": {"10"}, "3": {"15"},
		"U": {"Jan", "Bob", "Smit"}, "verbose": {"false"},
	}
	tests := [][]string{
		split("./app:-U:Jan:-UBob:-U:Smit:--no-verbose:5:10:15"),
		split("./app:5:10:-UJan:--no-verbose:-U:Bob:-U:Smit:15"),
	}

	am := argMap{}
	for _, test := range tests {
		if err := am.parse(test, flags); err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(am.asFlat(), expected) {
			t.Errorf("expected %v but %v", expected, am)
		}
	}
}

// TestParseErrors tests parse methods with errors.
func TestParseErrors(t *testing.T) {
	split := func(str string) []string { return strings.Split(str, ":") }
	flags := map[string]int{"U": 1, "verbose": 1}
	tests := [][]string{
		split("./app:-dU:Jan:-UBob:-U:Smit:--no-verbose"), // -d
		split("./app:-UJan:--users=Bob,Smit"),             // --users
		split("./app:-UJan:--no-users=Bob,Smit"),          // --no-users
	}

	am := argMap{}
	for _, test := range tests {
		if err := am.parse(test, flags); err == nil {
			t.Error("an error on a non-existent flag is expected")
		}
	}
}

// TestPositional tests positional method.
func TestPositional(t *testing.T) {
	split := func(str string) []string { return strings.Split(str, ":") }
	flags := map[string]int{"d": 1, "U": 1, "g": 1, "verbose": 1}
	expected := []string{"5", "10", "15"} // ignore app name "./app"
	tests := [][]string{
		split("./app:-dU:Jan,Bob:--no-verbose:-g:Hello, world:5:10:15"),
		split("./app:-d:-U:Jan,Bob:--no-verbose:-g:Hello, world:--:5:10:15"),
		split("./app:-dU:Jan,Bob:--verbose:false:-g:Hello, world:5:10:15"),
		split("./app:-dUJan,Bob:--verbose=false:-gHello, world:5:10:15"),
		split("./app:5:10:15:-dUJan,Bob:--verbose:false:-gHello, world"),
		split("./app:5:10:15:-UJan,Bob:--no-verbose:-gHello, world:-d:true"),
		split("./app:5:-UJan,Bob:--verbose:false:-gHello, world:-d:--:10:15"),
	}

	for i, test := range tests {
		am := argMap{}
		if err := am.parse(test, flags); err != nil {
			t.Error(err)
		}

		// Check positional arguments without app name.
		if r := am.posValues(); !reflect.DeepEqual(r, expected) {
			t.Errorf("%d test, expected %v but %v", i, expected, r)
		}

		// Check app name.
		value, _ := am.flagValue("0", "", "", "")
		if r := strings.Join(value, ""); r != tests[0][0] {
			t.Errorf("%d test, expected %v but %v", i, tests[0][0], r)
		}
	}
}

// TestValue tests value method.
func TestValue(t *testing.T) {
	split := func(str string) []string { return strings.Split(str, ":") }
	flags := map[string]int{"d": 1, "U": 1, "g": 1, "verbose": 1, "users": 1}
	expected := []string{"Jan,Bob"}
	tests := [][]string{
		// Short flag for users.
		split("./app:-dU:Jan,Bob:--no-verbose:-g:Hello, world:5:10:15"),
		split("./app:-d:-U:Jan,Bob:--no-verbose:-g:Hello, world:--:5:10:15"),
		split("./app:-dU:Jan,Bob:--verbose:false:-g:Hello, world:5:10:15"),

		// Long flag for users.
		split("./app:--users=Jan,Bob:--verbose=false:-gHello, world:5:10:15"),
		split("./app:5:10:15:--users:Jan,Bob:--verbose:false:-gHello, world"),

		// Default value.
		split("./app:5:10:15:--verbose:false:-gHello, world:-d:true"),
		split("./app:5:--verbose:false:-gHello, world:-d:--:10:15"),
	}

	am := argMap{}
	for _, test := range tests {
		if err := am.parse(test, flags); err != nil {
			t.Error(err)
		}

		r, _ := am.flagValue("U", "users", "Jan,Bob", "")
		if !reflect.DeepEqual(r, expected) {
			t.Errorf("expected %v but %v", expected, r)
		}
	}
}

// TestSplit tests split method.
func TestSplit(t *testing.T) {
	split := func(str string) []string { return strings.Split(str, ":") }
	equal := func(a, b []string) bool {
		am := make(map[string]bool, len(a))
		for _, key := range a {
			am[key] = true
		}

		bm := make(map[string]bool, len(b))
		for _, key := range b {
			bm[key] = true
		}

		return reflect.DeepEqual(am, bm)
	}

	flags := map[string]int{"d": 1, "U": 1, "g": 1, "verbose": 1, "users": 1}
	expected := []string{"Jan", "Bob", "Roy"}
	tests := [][]string{
		split("./app:-U:Jan:--users:Bob:--users=Roy"),
		split("./app:--users:Jan:-U:Bob:-URoy"),
	}

	am := argMap{}
	for i, test := range tests {
		if err := am.parse(test, flags); err != nil {
			t.Error(err)
		}

		r, _ := am.flagValue("U", "users", "Jan,Bob,Roy", "")
		if !equal(r, expected) {
			t.Errorf("test %d, expected %v but %v", i, expected, r)
		}
	}
}

// TestSplitFolding tests split method with folding data.
func TestSplitFolding(t *testing.T) {
	split := func(str string) []string { return strings.Split(str, ":") }
	flags := map[string]int{"d": 1, "U": 1, "g": 1, "verbose": 1, "users": 1}
	expected := []string{"Jan", "John,Bob", "Roy"}
	tests := [][]string{
		split("./app:-U:Jan:-U:John,Bob:-URoy"),
		split("./app:--users:Jan:-UJohn,Bob:-URoy"),
		split("./app:--users=Jan:-U:John,Bob:-URoy"),
	}

	am := argMap{}
	for i, test := range tests {
		if err := am.parse(test, flags); err != nil {
			t.Error(err)
		}

		r, _ := am.flagValue("U", "users", "Jan,Bob,Roy", "")
		if !reflect.DeepEqual(r, expected) {
			t.Errorf("test %d, expected %v but %v", i, expected, r)
		}
	}
}
