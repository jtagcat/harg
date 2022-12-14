# `harg`

[![Go Reference](https://pkg.go.dev/badge/github.com/jtagcat/hcli/harg.svg)](https://pkg.go.dev/github.com/jtagcat/hcli/harg)

GNU-compatible feature-complete local-scoped Go argument parsing. See [`FORMAT.md`](FORMAT.md) for full specification.

***

`harg` be a play on [git.meatballhat.com/x/argh](https://git.meatballhat.com/x/argh)! And more! Be fer 'human-friendly arguments' and 'harrrggghh'! Finally somethin' splendid fer do' Go arguments! 🏴‍☠️

## Next up:
- [`urfave/cli@v3`](https://github.com/urfave/cli)?

### Niceties:
- Definition-based shell completions
- `hyaml`: `yaml`?
- ~~Code generation?~~

## Terminology
`os.Args` is composed of the following:
 - program name (`harg`)
 - short options (`-v`, `-o something`)
 - long options (`--foo bar`, `--foo=bar`)
 - arguments (leftovers, `hello`, `world`)
 - chokeReturn (the found choke and everything after it, see [`Parse()`](parse.go) or [`FORMAT.md`](FORMAT.md) for more)

Non-boolean options have two parts: key (`foo`) and value (`bar`).
### Code flow
1. [`definition.go`](definition.go): definition structs
1. [`parse.go`](parse.go): main routine, splits to short/long option
1. [`parse_option.go`](parse_option.go): short and long option parsing
1. [`option_parse.go`](option_parse.go): parsing values to definitions
1. [`option_set.go`](option_set.go): typed structs
1. [`option_get.go`](option_get.go): typed structs, public functions for retrieving values.
