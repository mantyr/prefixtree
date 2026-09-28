package prefixtree_test

import (
	"fmt"

	"github.com/mantyr/prefixtree/v2"
)

// View показывает устройство дерева: [статика], :параметр, *catchall и значение
func ExampleView() {
	tree := prefixtree.New()
	for _, path := range []string{
		"/path/:dir/123",
		"/path/:dir/*filepath",
		"/path/user_:user",
		"/id/:id",
		"/id:id",
	} {
		if err := tree.SetString(path, path); err != nil {
			fmt.Println(err)
		}
	}
	for _, item := range prefixtree.View(tree) {
		fmt.Println(item)
	}
	// Output:
	// ^[/][path/]:dir[/][123]=/path/:dir/123
	// ^[/][path/]:dir[/]*filepath=/path/:dir/*filepath
	// ^[/][path/][user_]:user=/path/user_:user
	// ^[/][id][/]:id=/id/:id
	// ^[/][id]:id=/id:id
}
