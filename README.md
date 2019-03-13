# Golang Prefix Tree

[![Build Status](https://travis-ci.org/mantyr/prefixtree.svg?branch=master)](https://travis-ci.org/mantyr/prefixtree)
[![GoDoc](https://godoc.org/github.com/mantyr/prefixtree?status.png)](http://godoc.org/github.com/mantyr/prefixtree)
[![Go Report Card](https://goreportcard.com/badge/github.com/mantyr/prefixtree?v=1)][goreport]
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE.md)

This not stable version

## Installation

    $ go get github.com/mantyr/prefixtree

## Example
```GO
package main

import (
    "github.com/mantyr/prefixtree"
)
func main() {
    router := prefixtree.New("", prefixtree.Root)
    err := router.Set("/path/:dir/*filename", "string")
    err = router.Set("/user_:user", 123)
    data, params, err := router.Get("/path/123/file.zip")
}
```

## Author

[Oleg Shevelev][mantyr]

[mantyr]: https://github.com/mantyr

[build_status]: https://travis-ci.org/mantyr/prefixtree
[godoc]:        http://godoc.org/github.com/mantyr/prefixtree
[goreport]:     https://goreportcard.com/report/github.com/mantyr/prefixtree
