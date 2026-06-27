# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.1.0] - 2026-06-27

### Performance
- Cache the parsed struct specification (field tags, validation and the
  flag maps) per type, so the reflection and tag parsing run once per struct
  type instead of on every call. Repeated parsing of the same type is ~50%
  faster on small structs with fewer allocations; no API or behaviour change.

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