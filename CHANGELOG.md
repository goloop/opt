# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.4.1] - 2026-10-18

Preserve literal separators and reject negated non-boolean options.

## [2.4.0] - 2026-07-11

### Changed
- An absent flag no longer clobbers a pre-filled field, mirroring how
  encoding/json leaves fields missing from the input untouched; a `def`
  tag applies only while the field still holds its zero value. This makes
  the usual layering work: defaults, then environment, then flags.

## [2.3.0]

Correctness fixes across the reflection and parsing layers. A few behaviour
changes are noted below.

### Added
- Duplicate option names across fields are now detected and reported as a
  struct-configuration panic instead of silently sharing one value.

### Fixed
- An unexported struct field no longer becomes a flag or makes every parse fail;
  it is skipped, as `encoding/json` does.
- A nil pointer field (e.g. `Port *int`) is auto-allocated when a value is
  provided and left nil when the option is absent, instead of failing on every
  call with a leaked reflect error.
- A parsed list now replaces the field's contents instead of appending to it, so
  a pre-populated slice or a second `Unmarshal` no longer accumulates values.
- `Inf` / `-Inf` is accepted for a `float32` field (it is representable) rather
  than reported as an overflow.
- An array overflow records its error and keeps going, so the remaining fields
  and errors are still reported via `errors.Join`.
- A real empty positional argument (`app ""`) is kept, consistently with a
  non-empty one.
- The `-?` doc key and positional order keys are no longer accepted as
  command-line flags.
- The help text reports the true positional-array capacity (no off-by-one) and
  uses correct grammar ("takes one positional argument" / "takes up to N
  positional arguments").

### Changed
- **Behaviour:** a non-boolean option given without a value (e.g. `--host`
  followed by another flag) is now a usage error instead of being silently set
  to `"true"`.
- **Behaviour:** a lone `-` is treated as a positional operand (POSIX), not an
  invalid option.
- **Behaviour:** `Unmarshal`/`UnmarshalArgs` returns an error for a nil or
  non-pointer `obj` instead of panicking; a panic is now reserved for struct-tag
  misconfiguration.

## [2.1.0] - 2026-06-27

### Performance
- Cache the parsed struct specification (field tags, validation and the
  flag maps) per type, so the reflection and tag parsing run once per struct
  type instead of on every call. Repeated parsing of the same type is ~50%
  faster on small structs with fewer allocations; no API or behaviour change.
- Cache the generated help text per type. The help text depends only on the
  struct definition, so a struct with a `opt:"?"` field no longer regenerates
  it on every call — repeated parsing of such structs is ~3x faster with about
  half the allocations.

## [2.0.0] - 2026-06-27

### Changed (breaking)
- Module path is now `github.com/goloop/opt/v2`; import it as
  `github.com/goloop/opt/v2`.
- Boolean flags are pure switches (getopt-style) and never consume the next
  token. `--verbose false` now sets verbose to true and leaves `false` as a
  positional argument; use `--no-verbose` or `--verbose=false` to set false.
- Boolean options accept only proper boolean literals (`true`, `false`, `1`,
  `0`, ...); a value like `--debug=2` is now an error instead of true.
- `--flag=` with an empty value now means empty/zero, not the `true` switch.
- `Unmarshal` reports every error via `errors.Join` instead of only the first.
- Minimum Go version is 1.24.

### Added
- `UnmarshalArgs(obj any, args []string)` to parse an arbitrary argument slice.
- `required` tag modifier, e.g. `opt:"host,required"` (encoding/json style).
- `time.Duration` fields parsed from human-readable values (`30s`, `1h30m`).
- Negative numbers as option values and positional arguments (`-5`, `-3.14`).
- `-p=80` short-flag value form, symmetric with `--port=80`.
- Runnable Example functions and fuzz tests for the parser.

### Fixed
- Panic when a long flag name contained the digit `0` (e.g. `ipv0`).
- Options help sorting used an invalid comparator; positional arguments in the
  help block are now sorted and counted numerically (`10` after `2`).
- `opt:"-"` now actually ignores the field instead of exposing it as a flag.
- Fixed-size arrays fill left to right instead of being silently zeroed when
  fewer values are given.

### Performance
- Replaced flag-name regexps with manual checks and added fast paths to value
  collection, reducing allocations on the parsing hot path.