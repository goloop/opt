package opt

import (
	"math"
	"testing"
)

// TestUnexportedFieldIgnored guards BUG-01: an unexported field must not turn
// into a flag or make every parse fail.
func TestUnexportedFieldIgnored(t *testing.T) {
	type Args struct {
		Host  string `opt:"host"`
		count int    // unexported service field
	}
	a := Args{}
	if errs := unmarshalOpt(&a, []string{"app", "--host", "x"}); len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if a.Host != "x" {
		t.Errorf("Host = %q, want x", a.Host)
	}
	_ = a.count
	// "--count" must be rejected: the unexported field is not a flag.
	if errs := unmarshalOpt(&Args{}, []string{"app", "--count", "5"}); len(errs) == 0 {
		t.Error("--count should be an invalid argument")
	}
}

// TestNilPointerFieldAllocates guards BUG-02: a nil pointer field is allocated
// when a value is provided and left nil when absent.
func TestNilPointerFieldAllocates(t *testing.T) {
	type Args struct {
		Port *int `opt:"p" alt:"port"`
	}
	a := Args{}
	if errs := unmarshalOpt(&a, []string{"app", "--port", "80"}); len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if a.Port == nil || *a.Port != 80 {
		t.Fatalf("Port = %v, want 80", a.Port)
	}

	b := Args{}
	if errs := unmarshalOpt(&b, []string{"app"}); len(errs) != 0 {
		t.Fatalf("absent pointer errored: %v", errs)
	}
	if b.Port != nil {
		t.Errorf("absent Port = %v, want nil", b.Port)
	}
}

// TestMissingValueIsError guards BUG-03: a non-bool option with no value is an
// error, not an implicit "true".
func TestMissingValueIsError(t *testing.T) {
	type Args struct {
		Host string `opt:"host"`
		Port int    `opt:"p" alt:"port"`
	}
	a := Args{}
	if errs := unmarshalOpt(&a, []string{"app", "--host", "--port", "80"}); len(errs) == 0 {
		t.Error("--host with no value should error")
	}
	if a.Host == "true" {
		t.Error(`Host was set to "true" for a missing value`)
	}
}

// TestSliceReplacesNotAppends guards BUG-05: parsing replaces slice contents
// and is idempotent across calls.
func TestSliceReplacesNotAppends(t *testing.T) {
	type Args struct {
		Users []string `opt:"U" alt:"users"`
	}
	a := Args{Users: []string{"pre"}}
	if errs := unmarshalOpt(&a, []string{"app", "-U", "one", "-U", "two"}); len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(a.Users) != 2 || a.Users[0] != "one" || a.Users[1] != "two" {
		t.Fatalf("Users = %v, want [one two] (replace, not append)", a.Users)
	}
	// A second Unmarshal must not accumulate.
	if errs := unmarshalOpt(&a, []string{"app", "-U", "three"}); len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(a.Users) != 1 || a.Users[0] != "three" {
		t.Errorf("Users after second call = %v, want [three]", a.Users)
	}
}

// TestDuplicateFlagPanics guards BUG-06: two fields with the same option name
// are a struct misconfiguration and panic.
func TestDuplicateFlagPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic for duplicate option name")
		}
	}()
	type Args struct {
		A string `opt:"host"`
		B string `opt:"host"`
	}
	_ = unmarshalOpt(&Args{}, []string{"app"})
}

// TestServiceKeysNotFlags guards BUG-07: "-?" is rejected and an order key is
// not accepted as a flag.
func TestServiceKeysNotFlags(t *testing.T) {
	type Args struct {
		Help string `opt:"?"`
		A    int    `opt:"1"`
	}
	if errs := unmarshalOpt(&Args{}, []string{"app", "-?"}); len(errs) == 0 {
		t.Error("-? should be an invalid argument")
	}
	if errs := unmarshalOpt(&Args{}, []string{"app", "--1", "5"}); len(errs) == 0 {
		t.Error("--1 should be an invalid argument")
	}
}

// TestLoneDashIsPositional guards BUG-08: a lone "-" is an operand, not a flag.
func TestLoneDashIsPositional(t *testing.T) {
	type Args struct {
		Verbose bool     `opt:"v"`
		Pos     []string `opt:"[]"`
	}
	a := Args{}
	if errs := unmarshalOpt(&a, []string{"app", "-v", "-"}); len(errs) != 0 {
		t.Fatalf("lone dash errored: %v", errs)
	}
	if len(a.Pos) != 1 || a.Pos[0] != "-" {
		t.Errorf("Pos = %v, want [-]", a.Pos)
	}
}

// TestInfForFloat32 guards BUG-09: Inf is representable in float32 and accepted.
func TestInfForFloat32(t *testing.T) {
	type Args struct {
		F float32 `opt:"f"`
	}
	a := Args{}
	if errs := unmarshalOpt(&a, []string{"app", "-f", "Inf"}); len(errs) != 0 {
		t.Fatalf("Inf for float32 errored: %v", errs)
	}
	if !math.IsInf(float64(a.F), 1) {
		t.Errorf("F = %v, want +Inf", a.F)
	}
}

// TestUnmarshalNilReturnsError guards BUG-10: a nil/non-pointer obj returns an
// error rather than panicking.
func TestUnmarshalNilReturnsError(t *testing.T) {
	if err := UnmarshalArgs(nil, []string{"app"}); err == nil {
		t.Error("UnmarshalArgs(nil) = nil, want error")
	}
}

// TestArrayOverflowKeepsCollecting guards BUG-11: an array overflow does not
// abort collection of other field errors.
func TestArrayOverflowKeepsCollecting(t *testing.T) {
	type Args struct {
		Pos [1]int `opt:"[]"`
		N   int    `opt:"n"`
	}
	a := Args{}
	// Two positionals overflow [1]int, and "-n bad" is a separate error.
	errs := unmarshalOpt(&a, []string{"app", "one", "two", "-n", "bad"})
	if len(errs) < 2 {
		t.Errorf("got %d errors %v, want at least 2 (overflow and bad -n)", len(errs), errs)
	}
}

// TestEmptyPositionalKept guards BUG-12: a real empty positional argument is
// kept, consistently with a non-empty one.
func TestEmptyPositionalKept(t *testing.T) {
	type Args struct {
		Pos []string `opt:"[]"`
	}
	a := Args{}
	if errs := unmarshalOpt(&a, []string{"app", ""}); len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(a.Pos) != 1 || a.Pos[0] != "" {
		t.Errorf("Pos = %#v, want [\"\"]", a.Pos)
	}
}
