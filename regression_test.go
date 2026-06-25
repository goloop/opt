package opt

import (
	"os"
	"reflect"
	"strings"
	"testing"
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
		os.Args = test.args
		a := edgeArgs{}
		if err := Unmarshal(&a); err != nil {
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
	os.Args = []string{"./app", "-p=80"}
	a := edgeArgs{}
	if err := Unmarshal(&a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Port != 80 {
		t.Errorf("expected Port=80, got %d", a.Port)
	}
}

// TestNegativePositional covers BUG-03: a negative number passed as a
// positional argument must be kept as a value, not parsed as a flag.
func TestNegativePositional(t *testing.T) {
	os.Args = []string{"./app", "-5", "10"}
	a := edgeArgs{}
	if err := Unmarshal(&a); err != nil {
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
	os.Args = []string{"./app", "--ipv0", "test"}
	a := struct {
		IPv0 string `opt:"ipv0"`
		Mode string `opt:"mode0"`
	}{}

	if err := Unmarshal(&a); err != nil {
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

	optText, _ := getOptionBlock(fcl, argMap{})
	ia := strings.Index(optText, "-a")
	im := strings.Index(optText, "-m")
	iz := strings.Index(optText, "-z")
	if !(ia >= 0 && ia < im && im < iz) {
		t.Errorf("expected order -a < -m < -z, got:\n%s", optText)
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

	optText, _ := getOptionBlock(fcl, argMap{})
	if strings.Index(optText, "zebra") > strings.Index(optText, "apple") {
		t.Errorf("expected -a,--zebra before -b,--apple, got:\n%s", optText)
	}
}
