# Golang Parameterized Prefix Tree

[![Build Status](https://travis-ci.org/mantyr/prefixtree.svg?branch=master)](https://travis-ci.org/mantyr/prefixtree)
[![GoDoc](https://godoc.org/github.com/mantyr/prefixtree?status.png)](http://godoc.org/github.com/mantyr/prefixtree)
[![Go Report Card](https://goreportcard.com/badge/github.com/mantyr/prefixtree?v=3)][goreport]
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE.md)

This stable version

## Priorities for the selection of values

1. Static Node

2. Params Node

3. CatchAll Node

## Restrictions

1. CatchAll can be only one and only at the insert end
```GO
root.SetString("/path/*filepath*other", "value")      // error
root.SetString("/path/*filepath/*other", "value")     // error
root.SetString("/path/*filepath/123/*other", "value") // error
root.SetString("/path/*filepath", "value")            // OK
```

2. CatchAll has the lowest priority
```GO
root.SetString("/path/123/123", "value1")   // OK
root.SetString("/path/:id/123", "value2")   // OK
root.SetString("/path/*filepath", "value3") // OK

root.GetString("/path/123/123") // value1
root.GetString("/path/234/123") // value2
root.GetString("/path/123/234") // value3
```

## Installation

    $ go get -u github.com/mantyr/prefixtree


## Example
```GO
package main

import (
	"github.com/mantyr/prefixtree"
)
func main() {
	root := prefixtree.New()
	root.SetString("/path/:dir/123", "value1")
	root.SetString("/path/:dir/*filepath", "value2")
	root.SetString("/path/user_:user", "value3")
	root.SetString("/id/:id", "value4")
	root.SetString("/id:id", "value5")
	root.SetString(":id/:name/123", "value6")
	root.SetString("/id:id", "value7")                 // error: path already in use
	root.SetString("/id:id2", "value8")

	value, err := root.GetString("/path/123/file.zip") 
	/*
		value.Path = "/path/123/file.zip"
		value.Value = "value2"
		value.Params["dir"] = "123"
		value.Params["filepath"] = "file.zip"
	*/
	items := View(root)
	/*
		items = []string{
			"^[/][path/]:dir[/][123]=value1"
			"^[/][path/]:dir[/]*filepath=value2"
			"^[/][path/][user_]:user=value3"
			"^[/][id][/]:id=value4"
			"^[/][id]:id=value5"
			"^[/][id]:id2=value8"
			"^:id[/]:name[/123]=value6"
		}
	*/
}
```

## Author

[Oleg Shevelev][mantyr]

[mantyr]: https://github.com/mantyr

[build_status]: https://travis-ci.org/mantyr/prefixtree
[godoc]:        http://godoc.org/github.com/mantyr/prefixtree
[goreport]:     https://goreportcard.com/report/github.com/mantyr/prefixtree
