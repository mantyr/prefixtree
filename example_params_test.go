package prefixtree_test

import (
	"fmt"

	"github.com/mantyr/prefixtree/v2"
)

func ExampleValue() {
	tree := prefixtree.New()
	fmt.Println(tree.SetString("/u/:slot/api/*path", "api"))

	value, _ := tree.GetString("/u/abc/api/users/1")
	fmt.Println(value.Value())
	fmt.Println(value.Path())
	fmt.Println(value.Params())
	// Output:
	// <nil>
	// api
	// /u/abc/api/users/1
	// [{slot abc} {path users/1}]
}

func ExampleParams_ByName() {
	tree := prefixtree.New()
	fmt.Println(tree.SetString("/u/:slot/api/*path", "api"))

	value, _ := tree.GetString("/u/abc/api/users/1")
	fmt.Println(value.Params().ByName("slot"))
	fmt.Println(value.Params().ByName("nope"))
	// Output:
	// <nil>
	// abc true
	//  false
}

// Map собирает параметры в map, это аллокация на каждый вызов
func ExampleParams_Map() {
	tree := prefixtree.New()
	fmt.Println(tree.SetString("/u/:slot/api/*path", "api"))

	value, _ := tree.GetString("/u/abc/api/users/1")
	fmt.Println(value.Params().Map())
	// Output:
	// <nil>
	// map[path:users/1 slot:abc]
}
