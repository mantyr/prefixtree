package prefixtree

import (
	"fmt"
)

func ExampleNew() {
	root := New()
	fmt.Println(
		root.SetString("/path/:dir/123", "value1"),
	)
	fmt.Println(
		root.SetString("/path/:dir/*filepath", "value2"),
	)
	fmt.Println(
		root.SetString("/path/user_:user", "value3"),
	)
	fmt.Println(
		root.SetString("/id/:id", "value4"),
	)
	fmt.Println(
		root.SetString("/id:id", "value5"),
	)
	fmt.Println(
		root.SetString(":id/:name/123", "value6"),
	)
	fmt.Println(
		root.SetString("/id:id", "value7"),
	)
	fmt.Println(
		root.SetString("/id:id2", "value8"),
	)
	value, err := root.GetString("/path/123/file.zip")
	fmt.Println(err)
	fmt.Println(string(value.Path))
	fmt.Println(value.String())
	fmt.Println(value.Value)
	items := View(root)
	for _, item := range items {
		fmt.Println(item)
	}
	// Output:
	// <nil>
	// <nil>
	// <nil>
	// <nil>
	// <nil>
	// <nil>
	// path already in use
	// <nil>
	// <nil>
	// /path/123/file.zip
	// *filepath
	// value2
	// ^[/][path/]:dir[/][123]=value1
	// ^[/][path/]:dir[/]*filepath=value2
	// ^[/][path/][user_]:user=value3
	// ^[/][id][/]:id=value4
	// ^[/][id]:id=value5
	// ^[/][id]:id2=value8
	// ^:id[/]:name[/123]=value6
}