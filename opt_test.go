package opt

import (
	"os"
	"strings"
	"testing"
)

type Args struct {
	Host     string `opt:"host" alt:"h" def:"localhost"`
	Port     int    `opt:"p" alt:"port" def:"8080"`
	UserName string // default opt == user-name
	Ignored  bool   `opt:"-"`
	Verbose  bool   `opt:"v" alt:"verb" def:"true"`
	Path     string `opt:"0"`
	PosSlcie []int  `opt:"[]"`
	PosArray [8]int `opt:"[]"`
	Help     string `opt:"?"`
}

// TestLongFlags tests Unmarshal with long flags.
func TestLongFlags(t *testing.T) {

	split := func(str string) []string { return strings.Split(str, ":") }
	tests := []struct {
		Args     []string
		Expected Args
	}{
		{
			// ./app --host=0.0.0.0 --port 80 --user-name John --no-verb
			split("./app:--host=0.0.0.0:--port:80:--user-name:John:--no-verb"),
			Args{
				Host:     "0.0.0.0",
				Port:     80,
				UserName: "John",
				Verbose:  false,
			},
		},
		{
			// ./app --HOST localhost --User-Name="John Smith"
			split("./app:--HOST:localhost:--User-Name:John Smith"),
			Args{
				Host:     "localhost",
				Port:     8080,
				UserName: "John Smith",
				Verbose:  true,
			},
		},
		{
			// ./app --port 80 --verb=false
			split("./app:--port:80:--verb=false"),
			Args{
				Host:     "localhost",
				Port:     80,
				UserName: "",
				Verbose:  false,
			},
		},
	}

	for _, test := range tests {
		os.Args = test.Args
		args := Args{}
		if err := Unmarshal(&args); err != nil {
			t.Error(err)
		}

		if args.Host != test.Expected.Host {
			t.Errorf("expected %s but %s", test.Expected.Host, args.Host)
		}

		if args.Port != test.Expected.Port {
			t.Errorf("expected %d but %d", test.Expected.Port, args.Port)
		}

		if args.UserName != test.Expected.UserName {
			t.Errorf(
				"expected %s but %s",
				test.Expected.UserName,
				args.UserName,
			)
		}

		if args.Verbose != test.Expected.Verbose {
			t.Errorf(
				"expected %v but %v",
				test.Expected.Verbose,
				args.Verbose,
			)
		}
	}
}

// TestShortFlags tests Unmarshal with short flags.
func TestShortFlags(t *testing.T) {
	args := Args{}
	split := func(str string) []string { return strings.Split(str, ":") }
	tests := []struct {
		Args     []string
		Expected Args
	}{
		{
			// Normal command line arguments:
			// ./app -h 127.0.0.1 -p 80
			split("./app:-h:127.0.0.1:-p80"),
			Args{Host: "127.0.0.1", Port: 80},
		},
		{
			//
			// ./app --HOST localhost --User-Name="John Smith"
			split("./app:--HOST:localhost:--User-Name:John Smith"),
			Args{Host: "localhost", Port: 8080, UserName: "John Smith"},
		},
		// {
		// 	// Default value for Host and UserName.
		// 	// ./app --port 80
		// 	split("./app:--port:80"),
		// 	Args{Host: "localhost", Port: 80, UserName: ""},
		// },
	}

	for _, test := range tests {
		os.Args = test.Args
		if err := Unmarshal(&args); err != nil {
			t.Error(err)
		}

		if args.Host != test.Expected.Host {
			t.Errorf("expected %s but %s", test.Expected.Host, args.Host)
		}

		if args.Port != test.Expected.Port {
			t.Errorf("expected %d but %d", test.Expected.Port, args.Port)
		}
	}
}
