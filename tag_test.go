package opt

import "testing"

// TestGetTagValues tests getTagValues function.
func TestGetTagValues(t *testing.T) {
	type test struct {
		value  string
		result [4]string
	}

	var tests = []test{
		{"a,b,c", [4]string{"a", "b", "c", ""}},
		{"a,,c", [4]string{"a", "", "c", ""}},
		{"a,,", [4]string{"a", "", "", ""}},
		{"a,(b,c),d,e,f", [4]string{"a", "(b,c)", "d", "e,f"}},
	}

	for i, s := range tests {
		if r := getTagValues(s.value); r != s.result {
			t.Errorf("test %d is failed, expected %v but %v", i, s.result, r)
		}
	}
}

// TestGetTagSample tests getTagSample function.
func TestGetTagSample(t *testing.T) {
	type test struct {
		value string
		args  *tagSample
	}

	var tests = []test{
		{"a,b,c", &tagSample{"a", "b", "c", ""}},
		{",b,c", &tagSample{"", "b", "c", ""}},
		{"-,b,c", &tagSample{"-", "b", "c", ""}},
		{"a,,c", &tagSample{"a", "", "c", ""}},
		{",,c", &tagSample{"", "default", "c", ""}},
		{"a", &tagSample{"a", "", "", ""}},
		{"a,b,\"a, b, c\"", &tagSample{"a", "b", "a, b, c", ""}},
		{"11", &tagSample{"11", "", "", ""}},
	}

	for i, s := range tests {
		if args := getTagSample(s.value, "default"); *args != *s.args {
			t.Errorf("test %d is failed, expected %v but %v",
				i, s.args, args)
		}
	}
}

// TestGetArgumentsMethods tests methods of the tagSample struct.
func TestGetTagArgsMethods(t *testing.T) {
	type test struct {
		value     string
		isValid   bool
		isIgnored bool
	}

	var tests = []test{
		{"a,bcd,c", true, false},
		{",bcd,c", true, false},
		{"-,bcd,c", true, true},
		{"a,,c", true, false},
		{"a", true, false},
		{"a,bcd,\"a, b, c\"", true, false},
		{"11", false, true},
	}

	for i, s := range tests {
		args := getTagSample(s.value, "default")
		if args.IsValid() != s.isValid {
			t.Errorf("[isValid] test %d is failed, expected %v but %v",
				i, s.isValid, args.IsValid())
		}

		if args.IsIgnored() != s.isIgnored {
			t.Errorf("[isIgnored] test %d is failed, expected %v but %v",
				i, s.isIgnored, args.IsIgnored())
		}
	}
}
