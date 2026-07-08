package opt

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// edgeArgs is a struct used by the parsing regression tests.
type edgeArgs struct {
	Num       int      `opt:"n" alt:"num"`
	Threshold int      `opt:"threshold"`
	Ratio     float64  `opt:"r" alt:"ratio"`
	Port      int      `opt:"p" alt:"port"`
	Pos       []string `opt:"[]"`
	Help      string   `opt:"?"`
}

// TestNegativeNumberValue covers BUG-01: a negative number that follows an
// option expecting a value must be read as that value, not rejected as a flag.
func TestNegativeNumberValue(t *testing.T) {
	split := func(s string) []string { return strings.Split(s, ":") }
	tests := []struct {
		args      []string
		num       int
		threshold int
		ratio     float64
	}{
		{split("./app:-n:-5"), -5, 0, 0},          // short flag
		{split("./app:--num:-5"), -5, 0, 0},       // long alt of -n
		{split("./app:--threshold:-5"), 0, -5, 0}, // long flag
		{split("./app:-r:-3.14"), 0, 0, -3.14},    // negative float
		{split("./app:--threshold=-5"), 0, -5, 0}, // control: "=" form
	}

	for i, test := range tests {
		a := edgeArgs{}
		if err := UnmarshalArgs(&a, test.args); err != nil {
			t.Errorf("%d: unexpected error: %v", i, err)
			continue
		}

		if a.Num != test.num || a.Threshold != test.threshold ||
			a.Ratio != test.ratio {
			t.Errorf(
				"%d: expected num=%d threshold=%d ratio=%g, "+
					"got num=%d threshold=%d ratio=%g",
				i, test.num, test.threshold, test.ratio,
				a.Num, a.Threshold, a.Ratio,
			)
		}
	}
}

// TestShortFlagEquals covers BUG-02: a short option may receive its value via
// "=" (e.g. "-p=80"), symmetrically with the long form "--port=80".
func TestShortFlagEquals(t *testing.T) {
	a := edgeArgs{}
	if err := UnmarshalArgs(&a, []string{"./app", "-p=80"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Port != 80 {
		t.Errorf("expected Port=80, got %d", a.Port)
	}
}

// TestNegativePositional covers BUG-03: a negative number passed as a
// positional argument must be kept as a value, not parsed as a flag.
func TestNegativePositional(t *testing.T) {
	a := edgeArgs{}
	if err := UnmarshalArgs(&a, []string{"./app", "-5", "10"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"-5", "10"}
	if !reflect.DeepEqual(a.Pos, want) {
		t.Errorf("expected Pos=%v, got %v", want, a.Pos)
	}
}

// TestLongFlagZeroDigit covers BUG-06: a long option name may contain the
// digit 0 (e.g. "ipv0", "mode0") without panicking.
func TestLongFlagZeroDigit(t *testing.T) {
	a := struct {
		IPv0 string `opt:"ipv0"`
		Mode string `opt:"mode0"`
	}{}

	if err := UnmarshalArgs(&a, []string{"./app", "--ipv0", "test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.IPv0 != "test" {
		t.Errorf("expected IPv0=%q, got %q", "test", a.IPv0)
	}
}

// TestOptionSortShortOnly covers BUG-07: options that have only a short flag
// must be sorted (the previous conjunction comparator never sorted them).
func TestOptionSortShortOnly(t *testing.T) {
	fcl := fieldCastList{
		&fieldCast{tagGroup: &tagGroup{shortFlag: "z", helpMsg: "z help"}},
		&fieldCast{tagGroup: &tagGroup{shortFlag: "a", helpMsg: "a help"}},
		&fieldCast{tagGroup: &tagGroup{shortFlag: "m", helpMsg: "m help"}},
	}

	optText, _ := getOptionBlock(fcl)
	ia := strings.Index(optText, "-a")
	im := strings.Index(optText, "-m")
	iz := strings.Index(optText, "-z")
	if !(ia >= 0 && ia < im && im < iz) {
		t.Errorf("expected order -a < -m < -z, got:\n%s", optText)
	}
}

// TestEmptyEqualsValue covers BUG-04: "--flag=" with an explicitly empty
// value means empty/zero, not the "true" switch and not the next token.
func TestEmptyEqualsValue(t *testing.T) {
	a := edgeArgs{}
	if err := UnmarshalArgs(&a, []string{"./app", "--port=", "10"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Port != 0 {
		t.Errorf("expected Port=0, got %d", a.Port)
	}

	// The trailing "10" must not be swallowed as the value of --port.
	if want := []string{"10"}; !reflect.DeepEqual(a.Pos, want) {
		t.Errorf("expected Pos=%v, got %v", want, a.Pos)
	}
}

// boolArgs is used by the boolean-grab regression tests.
type boolArgs struct {
	Verbose bool     `opt:"v" alt:"verbose"`
	Pos     []string `opt:"[]"`
}

// TestBoolDoesNotEatPositional covers BUG-05: a boolean option is a pure
// switch (Linux getopt style) and never consumes the following token, so a
// positional argument after a boolean flag is preserved. A boolean is set
// to false only via "--no-flag" or the "=" form.
func TestBoolDoesNotEatPositional(t *testing.T) {
	split := func(s string) []string { return strings.Split(s, ":") }
	tests := []struct {
		args    []string
		verbose bool
		pos     []string
	}{
		{split("./app:-v:5"), true, []string{"5"}},                // short + pos
		{split("./app:--verbose:5"), true, []string{"5"}},         // long + pos
		{split("./app:--verbose:false"), true, []string{"false"}}, // not eaten
		{split("./app:--verbose=false"), false, nil},              // "=" sets false
		{split("./app:--no-verbose"), false, nil},                 // no- sets false
		{split("./app:-v"), true, nil},                            // bare switch
	}

	for i, test := range tests {
		a := boolArgs{}
		if err := UnmarshalArgs(&a, test.args); err != nil {
			t.Errorf("%d: unexpected error: %v", i, err)
			continue
		}

		if a.Verbose != test.verbose {
			t.Errorf("%d: expected Verbose=%v, got %v",
				i, test.verbose, a.Verbose)
		}

		if !reflect.DeepEqual(a.Pos, test.pos) {
			t.Errorf("%d: expected Pos=%v, got %v", i, test.pos, a.Pos)
		}
	}
}

// TestPositionalHelpNumericSort covers BUG-08: positional arguments in the
// help block are sorted and counted numerically, so "10" comes after "2".
func TestPositionalHelpNumericSort(t *testing.T) {
	fcl := fieldCastList{
		&fieldCast{tagGroup: &tagGroup{shortFlag: "1", helpMsg: "first"}},
		&fieldCast{tagGroup: &tagGroup{shortFlag: "10", helpMsg: "tenth"}},
		&fieldCast{tagGroup: &tagGroup{shortFlag: "2", helpMsg: "second"}},
	}

	posText := getPositionalBlock(fcl, 0)
	if !strings.Contains(posText, "takes up to 10") {
		t.Errorf("expected count 10, got:\n%s", posText)
	}

	i1 := strings.Index(posText, "first")
	i2 := strings.Index(posText, "second")
	i10 := strings.Index(posText, "tenth")
	if !(i1 >= 0 && i1 < i2 && i2 < i10) {
		t.Errorf("expected order first < second < tenth, got:\n%s", posText)
	}
}

// TestUnmarshalJoinsErrors covers BUG-09: Unmarshal reports every error,
// not just the first one.
func TestUnmarshalJoinsErrors(t *testing.T) {
	a := struct {
		A int `opt:"a"`
		B int `opt:"b"`
	}{}

	err := UnmarshalArgs(&a, []string{"./app", "-a", "notnum", "-b", "alsonot"})
	if err == nil {
		t.Fatal("expected an error")
	}

	msg := err.Error()
	if !strings.Contains(msg, "notnum") || !strings.Contains(msg, "alsonot") {
		t.Errorf("expected both errors, got: %q", msg)
	}
}

// TestArrayPartialFill covers BUG-10: a fixed array is filled left to right;
// fewer values than its size leave the rest at zero (previously the whole
// array was silently zeroed), while more values than its size overflow.
func TestArrayPartialFill(t *testing.T) {
	a := struct {
		Pos [8]int `opt:"[]"`
	}{}
	if err := UnmarshalArgs(&a, []string{"./app", "1", "2", "3"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := ([8]int{1, 2, 3, 0, 0, 0, 0, 0}); a.Pos != want {
		t.Errorf("expected %v, got %v", want, a.Pos)
	}

	// More values than the array size is an overflow error.
	b := struct {
		Pos [3]int `opt:"[]"`
	}{}
	if err := UnmarshalArgs(&b, []string{"./app", "1", "2", "3", "4"}); err == nil {
		t.Error("expected an overflow error")
	}
}

// TestUnmarshalArgs covers the exported UnmarshalArgs entry point that parses
// an arbitrary argument slice instead of os.Args.
func TestUnmarshalArgs(t *testing.T) {
	var cfg struct {
		Host string `opt:"host" def:"localhost"`
		Port int    `opt:"p" alt:"port" def:"8080"`
	}

	args := []string{"app", "--host", "0.0.0.0", "-p", "80"}
	if err := UnmarshalArgs(&cfg, args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "0.0.0.0" || cfg.Port != 80 {
		t.Errorf("got host=%q port=%d", cfg.Host, cfg.Port)
	}
}

// TestRequiredModifier covers the "required" modifier: a required option must
// be present on the command line, otherwise Unmarshal reports an error.
func TestRequiredModifier(t *testing.T) {
	type cfg struct {
		Host string `opt:"host,required"`
		Port int    `opt:"p" def:"80"`
	}

	// Missing required option -> error.
	missing := cfg{}
	if err := UnmarshalArgs(&missing, []string{"app", "-p", "8080"}); err == nil {
		t.Error("expected an error for the missing required --host")
	}

	// Provided -> ok.
	ok := cfg{}
	if err := UnmarshalArgs(&ok, []string{"app", "--host", "x"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ok.Host != "x" {
		t.Errorf("expected Host=x, got %q", ok.Host)
	}
}

// TestRequiredPositional covers "required" on a positional list: at least one
// positional argument must be present.
func TestRequiredPositional(t *testing.T) {
	type cfg struct {
		Args []string `opt:"[],required"`
	}

	if err := UnmarshalArgs(&cfg{}, []string{"app"}); err == nil {
		t.Error("expected an error for missing required positional args")
	}

	c := cfg{}
	if err := UnmarshalArgs(&c, []string{"app", "x"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestInvalidModifierPanics covers an unknown modifier: it is a struct
// definition error, so Unmarshal panics (developer's mistake).
func TestInvalidModifierPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a panic for an unknown modifier")
		}
	}()

	UnmarshalArgs(&struct {
		X int `opt:"x,bogus"`
	}{}, []string{"app"})
}

// TestDuration covers time.Duration fields parsed from human-readable forms.
func TestDuration(t *testing.T) {
	type cfg struct {
		Timeout time.Duration `opt:"timeout" alt:"t"`
	}

	c := cfg{}
	if err := UnmarshalArgs(&c, []string{"app", "--timeout", "1h30m"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := 90 * time.Minute; c.Timeout != want {
		t.Errorf("expected %v, got %v", want, c.Timeout)
	}

	// An invalid duration is an error.
	if err := UnmarshalArgs(&cfg{}, []string{"app", "-t", "nope"}); err == nil {
		t.Error("expected an error for an invalid duration")
	}
}

// TestStrictBoolValue covers BUG-11: a non-boolean value for a boolean
// option (e.g. --debug=2) is an error, not silently true.
func TestStrictBoolValue(t *testing.T) {
	type cfg struct {
		Debug bool `opt:"debug" alt:"d"`
	}

	if err := UnmarshalArgs(&cfg{}, []string{"app", "--debug=2"}); err == nil {
		t.Error("expected an error for --debug=2")
	}

	c := cfg{}
	if err := UnmarshalArgs(&c, []string{"app", "--debug=true"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !c.Debug {
		t.Error("expected Debug=true")
	}
}

// TestIgnoreMarker covers BUG-13: a field tagged opt:"-" is ignored entirely
// and is not exposed as a flag derived from its name.
func TestIgnoreMarker(t *testing.T) {
	type cfg struct {
		Ignored bool   `opt:"-"`
		Host    string `opt:"host"`
	}

	// --ignored is not a declared flag, so it is an error.
	bad := cfg{}
	if err := UnmarshalArgs(&bad, []string{"app", "--ignored"}); err == nil {
		t.Error("expected an error: opt:\"-\" must not create a flag")
	}
	if bad.Ignored {
		t.Error("opt:\"-\" field must not be settable")
	}

	// Other fields still work while the ignored one is skipped.
	ok := cfg{}
	if err := UnmarshalArgs(&ok, []string{"app", "--host", "x"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ok.Host != "x" {
		t.Errorf("expected Host=x, got %q", ok.Host)
	}
}

// TestOptionSortMixed covers BUG-07: options are sorted by the short flag,
// so "-a,--zebra" must precede "-b,--apple".
func TestOptionSortMixed(t *testing.T) {
	fcl := fieldCastList{
		&fieldCast{tagGroup: &tagGroup{
			shortFlag: "b", longFlag: "apple", helpMsg: "apple help",
		}},
		&fieldCast{tagGroup: &tagGroup{
			shortFlag: "a", longFlag: "zebra", helpMsg: "zebra help",
		}},
	}

	optText, _ := getOptionBlock(fcl)
	if strings.Index(optText, "zebra") > strings.Index(optText, "apple") {
		t.Errorf("expected -a,--zebra before -b,--apple, got:\n%s", optText)
	}
}
