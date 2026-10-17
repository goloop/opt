package opt

import "testing"

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
