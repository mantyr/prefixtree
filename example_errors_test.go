package prefixtree_test

import (
	"fmt"

	"github.com/mantyr/prefixtree/v2"
)

func ExampleCode() {
	tree := prefixtree.New()
	errs := []error{
		tree.SetString("/u/:slot/api/*path", "api"),
		tree.SetString("/u/:slot/api/*path", "again"),
		tree.SetString("/u/:id/:id", "dup"),
		tree.SetString("/u?x", "bad"),
		tree.SetString("/x", nil),
		tree.DeleteString("/nope"),
		tree.ReplaceString("/u/:slot/api/*path", "/u/:slot/api/*path", "api"),
	}
	for _, err := range errs {
		switch prefixtree.Code(err) {
		case prefixtree.OK:
			fmt.Println("ok")
		case prefixtree.PathNotFound:
			fmt.Println("not found")
		case prefixtree.PathAlreadyExists:
			fmt.Println("already exists")
		case prefixtree.DuplicateParamInPath:
			fmt.Println("duplicate param")
		case prefixtree.InvalidPath:
			fmt.Println("invalid path")
		case prefixtree.EmptyValue:
			fmt.Println("empty value")
		case prefixtree.NotChanged:
			fmt.Println("not changed")
		default:
			fmt.Println("unexpected")
		}
	}
	// Output:
	// ok
	// already exists
	// duplicate param
	// invalid path
	// empty value
	// not found
	// not changed
}
