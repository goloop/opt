package opt

import (
	"net/url"
	"testing"
)

// Test structures
type simpleArgs struct {
	Host    string `opt:"host" alt:"H" def:"localhost" help:"host of the server"`
	Port    int    `opt:"port" alt:"p" def:"8080" help:"port of the server"`
	Debug   bool   `opt:"d" help:"debug mode"`
	Verbose bool   `def:"true" help:"enable verbose mode"`
}

type complexArgs struct {
	Host      string   `opt:"H" alt:"host" def:"localhost" help:"host of the server"`
	Port      int      `opt:"port" alt:"p" def:"8080" help:"port of the server"`
	Help      bool     `opt:"h" alt:"help" help:"show application usage information"`
	Debug     bool     `opt:"d" help:"debug mode"`
	Verbose   bool     `def:"true" help:"enable verbose mode"`
	Configs   []string `opt:"c" sep:","`
	ServerURL url.URL  `opt:"U" help:"URL to the server"`
	Doc       string   `opt:"?"`
	Pos       []int    `opt:"[]"`
}

func BenchmarkUnmarshalSimple(b *testing.B) {
	args := []string{
		"./app",
		"--host=127.0.0.1",
		"-p8080",
		"--debug",
		"--no-verbose",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg simpleArgs
		if errs := unmarshalOpt(&cfg, args); errs != nil {
			b.Fatal(errs[0])
		}
	}
}

func BenchmarkUnmarshalComplex(b *testing.B) {
	args := []string{
		"./app",
		"-H", "127.0.0.1",
		"--port=8080",
		"-d",
		"--no-verbose",
		"-c", "config1.yaml,config2.yaml",
		"-U", "http://localhost:8080",
		"10",
		"20",
		"30",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var cfg complexArgs
		if errs := unmarshalOpt(&cfg, args); errs != nil {
			b.Fatal(errs[0])
		}
	}
}

func BenchmarkArgMapParse(b *testing.B) {
	args := []string{
		"./app",
		"-dUJack",
		"-U", "Bob",
		"--user=Roy",
		"--no-verbose",
		"-g", "L R",
		"5",
		"10",
		"15",
	}

	flags := map[string]int{
		"d":       1,
		"U":       1,
		"user":    1,
		"verbose": 1,
		"g":       1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am := argMap{}
		if err := am.parse(args, flags); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetTagGroup(b *testing.B) {
	cases := []struct {
		fieldName    string
		optTagValue  string
		altTagValue  string
		defTagValue  string
		sepTagValue  string
		helpTagValue string
	}{
		{"Host", "H", "host", "localhost", "", "host address"},
		{"Port", "port", "p", "8080", "", "port number"},
		{"Debug", "d", "", "", "", "debug mode"},
		{"Configs", "c", "", "", ",", "config files"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range cases {
			_, err := getTagGroup(
				c.fieldName,
				c.optTagValue,
				c.altTagValue,
				c.defTagValue,
				c.sepTagValue,
				c.helpTagValue,
			)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
