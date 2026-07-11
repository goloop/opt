package opt

import "testing"

// The usual configuration layering is defaults, then environment, then
// flags. Unmarshal must behave like encoding/json here: an absent flag
// leaves a pre-filled field untouched, and a def tag applies only while
// the field still holds its zero value.
func TestAbsentFlagPreservesPrefilled(t *testing.T) {
	type cfg struct {
		DB   string `opt:"db"`
		Addr string `opt:"addr" def:":8090"`
		Port int    `opt:"port" def:"8080"`
	}

	// Pre-filled by an earlier stage (for example the environment).
	c := cfg{DB: "from-env", Addr: ":9000"}
	if err := UnmarshalArgs(&c, []string{"app"}); err != nil {
		t.Fatal(err)
	}
	if c.DB != "from-env" {
		t.Errorf("DB = %q, want the pre-filled value kept", c.DB)
	}
	if c.Addr != ":9000" {
		t.Errorf("Addr = %q, want the pre-filled value to beat def", c.Addr)
	}
	if c.Port != 8080 {
		t.Errorf("Port = %d, want def on a zero field", c.Port)
	}

	// A flag still overrides everything.
	if err := UnmarshalArgs(&c, []string{"app", "--db", "from-flag"}); err != nil {
		t.Fatal(err)
	}
	if c.DB != "from-flag" {
		t.Errorf("DB = %q, want the flag to win", c.DB)
	}
}
