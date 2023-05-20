# days

`days` is a command-line tool for calculating the number of days between given dates. It does one thing and does it well.

Google Search already does this, but I don't use Google, and would prefer to have a simple offline tool to solve this problem.

`days` has no dependencies beyond the Go standard library.

## Usage

Some example commands:

```
days until june 16
days since feb 23
days since 'february 23'
days since feb 23 2004
days from jan 3 march 3
days from jan 3 to march 3
days from jan 3 2004 march 3 2006
days from 'jan 3 2004' 'march 3 2006'
days from jan 3 march 3 2030
days from jan 3 2004 march 3
```

The extra day during leap years is taken into account.

All dates are considered to be in your local timezone, but the timezone used can be changed by setting the standard `TZ` environment variable.

## Install

Right now `days` is just a personal tool that I've put up on GitHub in case anyone else would like to use it. As such I am not releasing any official binaries for now. You can install `days` by building from source with `go build`.

## License

`days` is licensed under the GPL v3.0. See the [LICENSE](./LICENSE) file for details.
