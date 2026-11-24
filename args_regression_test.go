package opt

import "testing"

func TestShortFlagAttachedValue(t *testing.T) {
	for _, tc := range []struct {
		name           string
		args           []string
		path           string
		verbose, debug bool
	}{
		{"registered flag", []string{"app", "-vpd"}, "d", true, false},
		{"word beginning with flag", []string{"app", "-vpdraft"}, "draft", true, false},
		{"without boolean prefix", []string{"app", "-pd"}, "d", false, false},
		{"boolean group", []string{"app", "-vd"}, "", true, true},
		{"boolean group and value", []string{"app", "-vdpdraft"}, "draft", true, true},
		{"equals", []string{"app", "-vp=d"}, "d", true, false},
		{"separate value", []string{"app", "-vp", "d"}, "d", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var cfg struct {
				Verbose bool   `opt:"v"`
				Path    string `opt:"p"`
				Debug   bool   `opt:"d"`
			}
			if err := UnmarshalArgs(&cfg, tc.args); err != nil {
				t.Fatal(err)
			}
			if cfg.Path != tc.path || cfg.Verbose != tc.verbose || cfg.Debug != tc.debug {
				t.Fatalf("got %+v, want Path=%q Verbose=%t Debug=%t", cfg, tc.path, tc.verbose, tc.debug)
			}
		})
	}
}

func TestLiteralSeparatorAfterTerminator(t *testing.T) {
	var cfg struct {
		Args []string `opt:"[]"`
	}
	if err := UnmarshalArgs(&cfg, []string{"app", "--", "--", "-x"}); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Args) != 2 || cfg.Args[0] != "--" || cfg.Args[1] != "-x" {
		t.Fatalf("args = %q", cfg.Args)
	}
}

func TestNegatedStringFlagIsRejected(t *testing.T) {
	var cfg struct {
		Name string `opt:"name"`
	}
	if err := UnmarshalArgs(&cfg, []string{"app", "--no-name"}); err == nil {
		t.Fatal("accepted a negated string option")
	}
}
