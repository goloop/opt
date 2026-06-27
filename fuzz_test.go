package opt

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// FuzzParse feeds random argument strings to argMap.parse and checks the core
// invariants: it never panics, and every key it stores is either a declared
// flag or a positional index.
func FuzzParse(f *testing.F) {
	seeds := []string{
		"./app -d -p 80 --host x 5 10",
		"./app --no-verbose -- -5 10",
		"./app -p=80 -dUvalue --host=",
		"./app --threshold -5 -r -3.14",
		"---  -- = -dd ----verbose",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	flags := map[string]int{
		"d": 1, "p": 1, "host": 1, "verbose": 1, "U": 1,
		"r": 1, "threshold": 1,
	}
	bools := map[string]bool{"d": true, "verbose": true}

	f.Fuzz(func(t *testing.T, s string) {
		args := strings.Fields(s)

		am := argMap{}
		_ = am.parse(args, flags, bools) // must not panic

		for key := range am {
			if _, ok := flags[key]; ok {
				continue
			}
			if _, err := strconv.Atoi(key); err == nil {
				continue // positional index
			}
			t.Errorf("unexpected key %q stored by parse(%q)", key, s)
		}
	})
}

// FuzzSetValue feeds random strings to setValue for every supported scalar
// kind and checks that it never panics (panics are recovered into errors).
func FuzzSetValue(f *testing.F) {
	for _, s := range []string{"", "123", "-5", "3.14", "true", "30s", "http://x"} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		var (
			i   int
			u   uint
			fl  float64
			b   bool
			str string
			d   time.Duration
			uu  url.URL
		)

		targets := []any{&i, &u, &fl, &b, &str, &d, &uu}
		for _, target := range targets {
			elem := reflect.ValueOf(target).Elem()
			_ = setValue(elem, s) // must not panic
		}
	})
}
