package opt

import (
	"net/url"
	"reflect"
	"testing"
)

// Test structures
type simpleArgs struct {
	Host    string `opt:"host" alt:"H" def:"localhost" help:"host of the server"`
	Port    int    `opt:"p" alt:"port" def:"8080" help:"port of the server"`
	Debug   bool   `opt:"d" help:"debug mode"`
	Verbose bool   `opt:"verbose" help:"enable verbose mode"`
}

type complexArgs struct {
	Host      string   `opt:"H" alt:"host" def:"localhost" help:"host of the server"`
	Port      int      `opt:"port" alt:"p" def:"8080" help:"port of the server"`
	Debug     bool     `opt:"d" help:"debug mode"`
	Verbose   bool     `opt:"verbose" help:"enable verbose mode"`
	Configs   []string `opt:"c" sep:","`
	ServerURL url.URL  `opt:"U" help:"URL to the server"`
	Doc       string   `opt:"?"`
	Pos       []int    `opt:"[]"`
}

var (
	testURL     *url.URL
	testArgMap  argMap
	testTagGrp  *tagGroup
	benchResult error
)

func BenchmarkParseArgMap(b *testing.B) {
	args := []string{
		"./app",
		"-d",
		"-p", "8080",
		"--host", "localhost",
	}
	flags := map[string]int{
		"d":    1,
		"p":    1,
		"host": 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am := make(argMap)
		benchResult = am.parse(args, flags)
	}
}

func BenchmarkParseArgMapComplex(b *testing.B) {
	args := []string{
		"./app",
		"-d",
		"-U", "http://localhost:8080",
		"--host=localhost",
		"--verbose=false",
		"-c", "config1.yaml,config2.yaml",
		"5",
		"10",
		"15",
	}
	flags := map[string]int{
		"d":       1,
		"U":       1,
		"host":    1,
		"verbose": 1,
		"c":       1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am := make(argMap)
		benchResult = am.parse(args, flags)
	}
}

func BenchmarkUnmarshalSimple(b *testing.B) {
	args := []string{
		"./app",
		"--host=127.0.0.1",
		"-p", "8080",
		"-d",
		"--verbose=false",
	}

	var cfg simpleArgs
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		errs := unmarshalOpt(&cfg, args)
		if errs != nil {
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
		"--verbose=false",
		"-c", "config1.yaml,config2.yaml",
		"-U", "http://localhost:8080",
		"10",
		"20",
		"30",
	}

	var cfg complexArgs
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		errs := unmarshalOpt(&cfg, args)
		if errs != nil {
			b.Fatal(errs[0])
		}
	}
}

func BenchmarkGetTagGroup(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tg, err := getTagGroup(
				"Host",
				"h",
				"host",
				"localhost",
				"",
				"host address",
			)
			testTagGrp = &tg
			benchResult = err
		}
	})

	b.Run("WithSeparator", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tg, err := getTagGroup(
				"Configs",
				"c",
				"",
				"conf1.yaml,conf2.yaml",
				",",
				"config files",
			)
			testTagGrp = &tg
			benchResult = err
		}
	})
}

func BenchmarkSetValue(b *testing.B) {
	var str string
	strVal := reflect.ValueOf(&str).Elem()

	var num int64
	numVal := reflect.ValueOf(&num).Elem()

	var bl bool
	boolVal := reflect.ValueOf(&bl).Elem()

	b.Run("String", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			benchResult = setValue(strVal, "test")
		}
	})

	b.Run("Int64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			benchResult = setValue(numVal, "12345")
		}
	})

	b.Run("Bool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			benchResult = setValue(boolVal, "true")
		}
	})
}
