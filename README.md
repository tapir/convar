# convar - A Quake-like console implementation for games [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/tapir/convar)

### Features

* Type-safe
* `int`, `bool`, `float64`, `string` primitive types
* Special `func` type that accepts a single argument of any primitive type
* Variables can trigger a callback function when set/updated
* Everything is concurrent safe
* Case insensitive variable names
* Helper functions (see docs)
* Saving/loading variables to/from a file
* Simple command evaluation/execution
* Simple autocomplete

## Installation

```
go get -u github.com/tapir/convar
```

## Examples

Check out the `_examples` folder.
![screenshot](https://github.com/tapir/convar/blob/master/_examples/ebicon/screenshot.png?raw=true)