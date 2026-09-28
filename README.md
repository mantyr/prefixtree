# Golang Parameterized Prefix Tree

[![Go Reference](https://pkg.go.dev/badge/github.com/mantyr/prefixtree.svg)](https://pkg.go.dev/github.com/mantyr/prefixtree)
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

## Params

`Get` returns `Value` — an interface with `Value()`, `Path()` and `Params()`.
Params are a slice in path order: `ByName(name)` looks one up, `Map()` builds a map (allocates on every call).
Param values are substrings of the requested path, nothing is copied.

A miss costs zero allocations, a hit costs exactly one — the result behind the `Value` interface.

## Delete

`Delete` removes a value by its pattern, exactly as it was set — `Delete("/u/:slot/api/*path")`, not by a request path.
Param names must match: `/u/:id/api/*path` is a different pattern.
Nodes left empty are removed and split static nodes are merged back, so the tree ends up as if the pattern had never been set.

## Replace

`Replace(old, new, v)` swaps one pattern for another in one step; if the patterns are equal, only the value changes and the node keeps its place among siblings.
It either fails without touching the tree or does the whole thing: every check for `new` runs before `old` is removed, taking the removal into account (`/a/*x` → `/a/*y` is fine).

## Storage

`Tree` is the full tree interface (`Set`, `Delete`, `Replace`, `Get`, `Clone`), `TreeReadonly` has only `Get`.
`Storage` is a tree for concurrent reads: `Set`, `Delete` and `Replace` change a master tree under a mutex
and then publish its copy through an atomic pointer; `Tree()` returns that copy as `TreeReadonly`, readers never wait.
On error nothing is published; `Replace` with the same pattern and an equal value returns `NotChanged` and publishes nothing either.

```go
s := prefixtree.NewStorage()

err := s.Replace("/u/:slot/api/*path", "/u/:slot/v2/*path", handler)

value, err := s.Tree().GetString(r.URL.Path)
```

## Errors

```go
value, err := root.GetString("/u/abc/oauth/x")
switch prefixtree.Code(err) {
case prefixtree.OK:
case prefixtree.PathNotFound:
default:
}
```

| Code | Returned by |
| --- | --- |
| `OK` | no error |
| `PathNotFound` | `Get` — no value for the path; `Delete`, `Replace` — no such pattern |
| `PathAlreadyExists` | `Set`, `Replace` — the pattern is already taken |
| `InvalidPath` | `Set`, `Delete`, `Replace` — the pattern cannot be parsed |
| `EmptyValue` | `Set` — nil value |
| `DuplicateParamInPath` | `Set` — the same param name is used twice in one pattern |
| `NotChanged` | `Replace` — same pattern and the same value (same pointer), nothing was changed or published |
| `Internal` | broken tree invariant (bug) |
| `Unknown` | error is not from this package |

## Documentation

Runnable examples for `New`, `Tree`, `Storage`, `Params`, `Value`, `Code`, `View` and a small HTTP router
live in `example_*_test.go` and are shown on [pkg.go.dev](https://pkg.go.dev/github.com/mantyr/prefixtree/v2).

## Installation

    $ go get -u github.com/mantyr/prefixtree/v2


## Example
```GO
package main

import (
	"github.com/mantyr/prefixtree/v2"
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
		value.Path()   = "/path/123/file.zip"
		value.Value()  = "value2"
		value.Params() = prefixtree.Params{{Key: "dir", Value: "123"}, {Key: "filepath", Value: "file.zip"}}
		value.Params().ByName("dir") = "123", true
	*/
	items := prefixtree.View(root)
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
