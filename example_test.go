package opt_test

import (
	"fmt"
	"time"

	"github.com/goloop/opt/v2"
)

// Parse a typical set of options from an explicit argument slice.
func ExampleUnmarshalArgs() {
	var cfg struct {
		Host    string `opt:"host" def:"localhost"`
		Port    int    `opt:"p" alt:"port" def:"8080"`
		Verbose bool   `opt:"v" alt:"verbose"`
	}

	args := []string{"app", "--host", "0.0.0.0", "-p", "80", "-v"}
	if err := opt.UnmarshalArgs(&cfg, args); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%s:%d verbose=%t\n", cfg.Host, cfg.Port, cfg.Verbose)
	// Output: 0.0.0.0:80 verbose=true
}

// Positional arguments are collected into a list field tagged opt:"[]".
// Negative numbers are kept as values, not treated as flags.
func ExampleUnmarshalArgs_positional() {
	var cfg struct {
		Args []string `opt:"[]"`
	}

	if err := opt.UnmarshalArgs(&cfg, []string{"app", "-5", "10", "x"}); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cfg.Args)
	// Output: [-5 10 x]
}

// A time.Duration field accepts human-readable values like "1h30m".
func ExampleUnmarshalArgs_duration() {
	var cfg struct {
		Timeout time.Duration `opt:"timeout" alt:"t"`
	}

	if err := opt.UnmarshalArgs(&cfg, []string{"app", "-t", "1h30m"}); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(cfg.Timeout)
	// Output: 1h30m0s
}
