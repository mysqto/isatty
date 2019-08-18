# isatty

[![Godoc Reference](https://godoc.org/github.com/mysqto/isatty?status.svg)](http://godoc.org/github.com/mysqto/isatty)
[![Build Status](https://travis-ci.org/mysqto/isatty.svg?branch=master)](https://travis-ci.org/mysqto/isatty)
[![Coverage Status](https://coveralls.io/repos/github/mysqto/isatty/badge.svg?branch=master)](https://coveralls.io/github/mysqto/isatty?branch=master)
[![Go Report Card](https://goreportcard.com/badge/mysqto/isatty)](https://goreportcard.com/report/mysqto/isatty)

isatty for golang

## Usage

```go
package main

import (
	"fmt"
	"github.com/mysqto/isatty"
	"os"
)

func main() {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Println("Is Terminal")
	} else if isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		fmt.Println("Is Cygwin/MSYS2 Terminal")
	} else {
		fmt.Println("Is Not Terminal")
	}
}
```

## Installation

```bash
$go get github.com/mysqto/isatty
```

## License

MIT

## Author

Chen Lei (a.k.a mysqto)
Yasuhiro Matsumoto (a.k.a mattn)

## Thanks

* k-takata: base idea for IsCygwinTerminal

    https://github.com/k-takata/go-iscygpty
