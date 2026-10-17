[![deps.dev](https://img.shields.io/badge/deps.dev-insights-4c8dbc)](https://deps.dev/go/github.com%2Fgoloop%2Fopt%2Fv2) [![License](https://img.shields.io/badge/license-MIT-brightgreen)](https://github.com/goloop/opt/blob/master/LICENSE) [![License](https://img.shields.io/badge/godoc-YES-green)](https://pkg.go.dev/github.com/goloop/opt/v2) [![Stay with Ukraine](https://img.shields.io/static/v1?label=Stay%20with&message=Ukraine%20♥&color=ffD700&labelColor=0057B8&style=flat)](https://u24.gov.ua/)

# opt

`opt` maps command-line arguments onto a Go struct — the same way
`encoding/json` maps JSON onto a struct. You describe the options with struct
tags, call `Unmarshal`, and read the typed fields.

It follows POSIX/getopt conventions: long (`--host`) and short (`-p`) flags,
short-flag grouping (`-vp80`), `--flag value` and `--flag=value`, boolean
switches that never consume the next token, positional arguments on either side,
and `--` to separate trailing positionals. Supported field types include every
sized int/uint, float, `string`, `bool`, `time.Duration`, `url.URL`, and
pointers/slices/arrays of those.

## Features

- **Struct-tag driven** — `opt`, `alt`, `def`, `sep`, `help`, plus a `required`
  modifier (`opt:"host,required"`).
- **Long & short flags** — pair them via `opt`/`alt`; group short flags.
- **Boolean switches** — `--verbose`, `--no-verbose`, `--verbose=false`.
- **Positional arguments** — collect all with `opt:"[]"`, or pick one by index.
- **Lists** — repeat a flag or split one value with `sep`.
- **Auto-generated help** — capture it in an `opt:"?"` field.
- **`UnmarshalArgs`** — parse an arbitrary argument slice (great for tests).

## Installation

```shell
go get -u github.com/goloop/opt/v2
```

```go
import "github.com/goloop/opt/v2"
```

Requires Go 1.24 or newer.

## Quick start

```go
package main

import (
    "fmt"
    "log"
    "net/url"
    "os"

    "github.com/goloop/opt/v2"
)

type Args struct {
    Host    string   `opt:"H" alt:"host" def:"localhost" help:"host of the server"`
    Port    int      `opt:"port" alt:"p" def:"8080" help:"port of the server"`
    Help    bool     `opt:"h" alt:"help" help:"show application usage information"`
    Debug   bool     `opt:"d" help:"debug mode"`
    Verbose bool     `def:"true" help:"enable verbose mode"` // --verbose
    Configs []string `opt:"c" sep:","`

    ServerURL url.URL `opt:"U" help:"URL to the server"`

    Doc        string `opt:"?"`  // receives the generated help text
    Positional []int  `opt:"[]"` // receives positional arguments
}

func main() {
    var args Args
    if err := opt.Unmarshal(&args); err != nil {
        log.Fatal(err)
    }

    if args.Help {
        fmt.Println(args.Doc)
        os.Exit(0)
    }

    fmt.Println("Host:", args.Host, "Port:", args.Port, "Positional:", args.Positional)
}
```

All of these parse to the same result — flag order does not matter:

```shell
./app -H 0.0.0.0 --port=80 --no-verbose -U goloop.one -dc a.yaml,b.yaml 5 10 15
./app 5 --host 0.0.0.0 -p 80 --no-verbose -U goloop.one -c a.yaml -c b.yaml -d -- 10 15
./app -c a.yaml,b.yaml -p 80 -H 0.0.0.0 --verbose=false -dU goloop.one 5 10 15
```

## Struct tags

| Tag | Purpose |
|-----|---------|
| `opt`  | short or long flag name; `+ ,required` modifier; reserved: `-`, `?`, `[]`, index digits |
| `alt`  | the alternative name (short if `opt` is long, or vice versa) |
| `def`  | default value when the flag is absent |
| `sep`  | delimiter that splits a list value |
| `help` | short description shown in the generated help |

See **[DOC.md](DOC.md)** for the complete parsing rules — grouping, boolean
switches, positional handling, duplicate flags, the panic-vs-error contract and
recipes (Ukrainian: **[DOC.UK.md](DOC.UK.md)**).

## Documentation

- Full reference and recipes: [DOC.md](DOC.md) · [DOC.UK.md](DOC.UK.md)
- Package API: [pkg.go.dev/github.com/goloop/opt/v2](https://pkg.go.dev/github.com/goloop/opt/v2)
- Changes between versions: [CHANGELOG.md](CHANGELOG.md)

## Contributing

Contributions are welcome. Please run `go test ./...`, `go vet ./...` and
`gofmt -l .` before submitting a pull request.

## License

`opt` is released under the MIT License. See [LICENSE](LICENSE).

## Release v2.4.1

Preserve literal separators and reject negated non-boolean options.
