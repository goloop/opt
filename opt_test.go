package opt

import (
	"os"
	"strings"
	"testing"
)

// The testArgs is example of the test args.
type testArgs struct {
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
		Expected testArgs
	}{
		{
			// ./app --host=0.0.0.0 --port 80 --user-name John --no-verb
			split("./app:--host=0.0.0.0:--port:80:--user-name:John:--no-verb"),
			testArgs{
				Host:     "0.0.0.0",
				Port:     80,
				UserName: "John",
				Verbose:  false,
			},
		},
		{
			// ./app --HOST localhost --User-Name="John Smith"
			split("./app:--HOST:localhost:--User-Name:John Smith"),
			testArgs{
				Host:     "localhost",
				Port:     8080,
				UserName: "John Smith",
				Verbose:  true,
			},
		},
		{
			// ./app --port 80 --verb=false
			split("./app:--port:80:--verb=false"),
			testArgs{
				Host:     "localhost",
				Port:     80,
				UserName: "",
				Verbose:  false,
			},
		},
	}

	for _, test := range tests {
		os.Args = test.Args
		args := testArgs{}
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
	args := testArgs{}
	split := func(str string) []string { return strings.Split(str, ":") }
	tests := []struct {
		Args     []string
		Expected testArgs
	}{
		{
			// ./app -h 127.0.0.1 -p 80
			split("./app:-h:127.0.0.1:-p80"),
			testArgs{Host: "127.0.0.1", Port: 80},
		},
		{
			// ./app --HOST localhost --User-Name="John Smith"
			split("./app:--HOST:localhost:--User-Name:John Smith"),
			testArgs{Host: "localhost", Port: 8080, UserName: "John Smith"},
		},
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

// TestUnmarshalErrors tests Unmarshal with errors.
func TestUnmarshalErrors(t *testing.T) {
	var args = struct {
		Host string
	}{}

	os.Args = []string{"./apt", "-d", "-p"} // flags d and p doesn't exist
	if err := Unmarshal(&args); err == nil {
		t.Error("expected an error")
	}
}
