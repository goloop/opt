# opt — reference

The full reference for the `opt` package: the mental model, the two entry
points, the struct tags, the complete parsing rules, boolean switches, supported
types, the panic-vs-error contract and practical recipes.

Ukrainian version: **[DOC.UK.md](DOC.UK.md)**.

## Contents

- [Mental model](#mental-model)
- [Entry points](#entry-points)
- [Struct tags](#struct-tags)
- [Reserved opt values](#reserved-opt-values)
- [Supported types](#supported-types)
- [Parsing rules](#parsing-rules)
  - [Long flags](#long-flags)
  - [Short flags and grouping](#short-flags-and-grouping)
  - [Boolean switches](#boolean-switches)
  - [Positional arguments](#positional-arguments)
  - [Duplicate flags and lists](#duplicate-flags-and-lists)
- [Generated help](#generated-help)
- [Panic vs error](#panic-vs-error)
- [Recipes and tips](#recipes-and-tips)

## Mental model

`opt` maps command-line arguments onto a Go struct, the same way `encoding/json`
maps JSON onto a struct. You describe the options with struct tags, call
`Unmarshal`, and read the typed fields.

Two categories of argument exist:

- **Flags** change behaviour. They come in *long* (`--host`) and *short* (`-p`)
  forms, and carry a value or act as a switch.
- **Values** are data — passed positionally (`5 10 15`) or as a flag's value
  (`--host=localhost`).

The parser follows POSIX/getopt conventions: short flags can be grouped
(`-vp80`), long flags accept `--flag value` or `--flag=value`, `--` separates
trailing positional arguments from flags, and booleans are switches that never
consume the next token.

Because absent flags leave fields untouched (like fields missing from a JSON
document), the usual configuration layering falls out naturally: fill the
struct with defaults, overlay the environment, then call `Unmarshal` - each
stage only overrides what it actually names, and a `def` tag never beats a
value an earlier stage already set.

```go
import "github.com/goloop/opt/v2"
```

## Entry points

```go
func Unmarshal(obj any) error
func UnmarshalArgs(obj any, args []string) error
```

`Unmarshal` parses `os.Args` into the struct pointed to by `obj`.
`UnmarshalArgs` does the same for an arbitrary slice — the first element is
treated as the program path (like `os.Args[0]`), which makes it convenient for
tests, sub-commands and REPLs:

```go
var cfg Args
err := opt.UnmarshalArgs(&cfg, []string{"app", "--host", "x", "-p", "80"})
```

`Unmarshal` returns all parse errors joined with `errors.Join` (do not match on
the error text). It can also panic on a *programming* mistake — see
[Panic vs error](#panic-vs-error).

## Struct tags

| Tag | Purpose | Default |
|-----|---------|---------|
| `opt`  | short or long flag name; may carry comma-separated modifiers, e.g. `opt:"host,required"` | field name in kebab-case |
| `alt`  | the alternative name (short if `opt` is long, or vice versa) | none |
| `def`  | default applied when the flag is absent **and** the field still holds its zero value | field left untouched |
| `sep`  | delimiter that splits a list value | empty (one flag per item) |
| `help` | short description shown in generated help | none (hidden) |

When `opt` is omitted, the flag name is the field name converted to kebab-case
(`UserName` → `--user-name`). `opt` and `alt` may not both be long or both be
short — the pair must be one short and one long.

The only modifier currently supported is `required`: the option must appear on
the command line, otherwise `Unmarshal` returns an error.

```go
type Args struct {
    Host    string   `opt:"H" alt:"host" def:"localhost" help:"host of the server"`
    Port    int      `opt:"port" alt:"p" def:"8080" help:"port of the server"`
    Verbose bool     `def:"true" help:"enable verbose mode"` // --verbose
    Configs []string `opt:"c" sep:","`
    DSN     string   `opt:"dsn,required"`
}
```

## Reserved opt values

The `opt` tag has four reserved values:

| Value | Meaning |
|-------|---------|
| `-`     | ignore the field entirely |
| `?`     | receive the generated help text (field must be `string`) |
| `[]`    | receive the positional arguments (field must be a slice/array) |
| `0`,`1`,…,`N` | a specific positional argument by index (index `0` is the program path) |

```go
type Args struct {
    Doc        string `opt:"?"`  // filled with the help text
    FileName   string `opt:"1"`  // the first positional argument
    Positional []int  `opt:"[]"` // all positional arguments
}
```

## Supported types

`int`/`int8`…`int64`, `uint`/`uint8`…`uint64`, `float32`/`float64`, `string`,
`bool`, `time.Duration`, `url.URL`, pointers to any of these, and arrays or
slices of any of these (e.g. `*int`, `[]string`, `[2]*url.URL`). Negative
numbers are accepted as values and positional arguments.

Any other field type (`chan`, `map`, other structs) is a programming error and
triggers a panic.

## Parsing rules

### Long flags

A long flag has two dashes: `--host`. Names are latin letters/digits with
hyphens for word breaks (`--user-name`) and are **case-insensitive**
(`--host` == `--HOST`). A value follows via a space or `=`:
`--host localhost` == `--host=localhost`. Multi-word values need quotes:
`--user="Smith J."`.

### Short flags and grouping

A short flag has one dash: `-p`. Names are latin letters only and are
**case-sensitive** (`-h` ≠ `-H`). A value follows with or without a space:
`-p 8080` == `-p8080`.

Short flags can be grouped: `-vdp8080` is `-v -d -p 8080`. Watch for accidental
duplicates in a group — in `-vpv` the trailing `v` becomes the value of `-p`
(i.e. `-v -p v`), because once a value-taking flag is seen the rest of the group
is its value.

### Boolean switches

A boolean flag is a switch. Writing `--verbose` (or `-v`) sets it to true and it
**never** consumes the next token — exactly like a no-argument option in
`getopt`. Its absence leaves the default (false unless `def` says otherwise).

To set false explicitly, use the `no-` prefix or the `=` form:

| Input | `verbose` |
|-------|-----------|
| `--verbose`, `-v`            | true |
| `--no-verbose`              | false |
| `--verbose=false`, `-v=false` | false |
| `--verbose false`           | true, and `false` becomes a positional argument |

A boolean value must be a real boolean literal; `--debug=2` is an error, and
`--flag=` (empty value) means zero/false, not true.

### Positional arguments

Positional arguments are values that do not belong to a flag. Index `0` is the
program path (added automatically). They may appear on the left (before the
first flag) and on the right (after the last flag's value), even at once:

```
./app 5 10 --host localhost -d -- 15
```

If the last flag before the trailing positionals is a switch (takes no value),
separate the positionals with `--`, otherwise the first one is read as that
flag's value. `--` marks "everything after me is positional".

### Duplicate flags and lists

A flag mapped to a slice/array may be repeated, and its values accumulate:

```
./app -uJohn --user=Bob -u Roy   // []string{"John", "Bob", "Roy"}
```

With a `sep`, a list may also be passed as one value (`-A23,25,27` with
`sep:","`). Repeating a flag that is **not** a list is not an error — the last
value wins.

## Generated help

Fill a `string` field tagged `opt:"?"` to capture auto-generated help, printed
when your own `-h`/`--help` flag is set:

```
Options:
    -H, --host    host of the server;
    -d            debug mode;
    -h, --help    show application usage information;
    -p, --port    port of the server;
        --verbose enable verbose mode;
    -U            URL to the server.

Positional arguments:
The app takes an unlimited number of positional arguments.
```

Options with an empty `help` tag are omitted from this text. The help text is
generated even when parsing fails, so you can print it alongside an error.

## Panic vs error

The two failure modes are deliberately different:

- **Panic — a development problem.** The struct or tags are wrong: `obj` is not
  a pointer to a struct; the `opt:"?"` field is not a `string`; the `opt:"[]"`
  field is not a list; a field has an unsupported struct/pointer type (other
  than `url.URL`/`*url.URL`). These cannot be fixed at runtime, so they panic.

- **Error — a user problem.** The command line is invalid: an undeclared flag,
  a value that doesn't fit the field type, an overflowing array, a number out of
  range, or a missing `required` option. These are returned so you can show help
  and exit gracefully:

```go
switch err := opt.Unmarshal(&args); {
case err != nil:
    log.Fatalf("%v\n\nError: %v", args.Doc, err)
case args.Help:
    fmt.Println(args.Doc)
    os.Exit(0)
}
```

## Recipes and tips

**Short + long pair.** Put one form in `opt` and the other in `alt`
(`opt:"host" alt:"H"`); the parser accepts both `-H` and `--host`.

**Required options.** Add the `required` modifier (`opt:"dsn,required"`) so a
missing value becomes a clear error instead of a silent zero.

**Test without touching `os.Args`.** Drive `UnmarshalArgs` with a literal slice
(remember the leading program-path element) to unit-test parsing.

**Collect a list two ways.** Give list fields a `sep` so callers can either
repeat the flag (`-c a -c b`) or pass one delimited value (`-c a,b`).

**Trailing positionals after a switch.** End the command line with `--` before
positional values when the last flag is a boolean, so they aren't swallowed as
its value.

## Release v2.4.1

Preserve literal separators and reject negated non-boolean options.
