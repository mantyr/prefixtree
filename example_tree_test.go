package prefixtree_test

import (
	"fmt"

	"github.com/mantyr/prefixtree/v2"
)

func ExampleNew() {
	tree := prefixtree.New()
	fmt.Println(tree.SetString("/u/:slot/api/*path", "api"))

	value, err := tree.GetString("/u/abc/api/users/1")
	fmt.Println(err)
	fmt.Println(value.Value())
	fmt.Println(value.Params())
	// Output:
	// <nil>
	// <nil>
	// api
	// [{slot abc} {path users/1}]
}

// Статика важнее параметра, параметр важнее CatchAll, порядок добавления не важен
func ExampleTree_set() {
	tree := prefixtree.New()
	fmt.Println(tree.SetString("/files/*path", "catch-all"))
	fmt.Println(tree.SetString("/files/:name", "param"))
	fmt.Println(tree.SetString("/files/readme", "static"))
	fmt.Println(tree.SetString("/files/readme", "again"))

	for _, path := range []string{"/files/readme", "/files/logo.png", "/files/img/logo.png"} {
		value, _ := tree.GetString(path)
		fmt.Println(path, "->", value.Value(), value.Params())
	}
	// Output:
	// <nil>
	// <nil>
	// <nil>
	// path already in use
	// /files/readme -> static []
	// /files/logo.png -> param [{name logo.png}]
	// /files/img/logo.png -> catch-all [{path img/logo.png}]
}

// Если ветка не дошла до конца, поиск откатывается и пробует следующую
func ExampleTree_get() {
	tree := prefixtree.New()
	fmt.Println(tree.SetString("/u/admin/api/*path", "admin"))
	fmt.Println(tree.SetString("/u/:slot/oauth/*path", "oauth"))

	for _, path := range []string{"/u/admin/api/x", "/u/admin/oauth/x", "/u/admin/other/x"} {
		value, err := tree.GetString(path)
		if err != nil {
			fmt.Println(path, "->", err)
			continue
		}
		fmt.Println(path, "->", value.Value(), value.Params())
	}
	// Output:
	// <nil>
	// <nil>
	// /u/admin/api/x -> admin [{path x}]
	// /u/admin/oauth/x -> oauth [{slot admin} {path x}]
	// /u/admin/other/x -> not found
}

// Удаляется шаблон целиком, имена параметров должны совпадать
func ExampleTree_delete() {
	tree := prefixtree.New()
	fmt.Println(tree.SetString("/u/:slot/api/*path", "api"))

	fmt.Println(tree.DeleteString("/u/:id/api/*path"))
	fmt.Println(tree.DeleteString("/u/:slot/api/*path"))

	_, err := tree.GetString("/u/abc/api/x")
	fmt.Println(err)
	// Output:
	// <nil>
	// not found
	// <nil>
	// not found
}

// Replace меняет шаблон одним шагом: либо ошибка и дерево не тронуто, либо всё сделано
func ExampleTree_replace() {
	tree := prefixtree.New()
	fmt.Println(tree.SetString("/u/:slot/api/*path", "v1"))

	// другой шаблон: старый удаляется, новый добавляется
	fmt.Println(tree.ReplaceString("/u/:slot/api/*path", "/u/:slot/v2/*path", "v2"))

	// тот же шаблон с новым значением: меняется только значение
	fmt.Println(tree.ReplaceString("/u/:slot/v2/*path", "/u/:slot/v2/*path", "v3"))

	// тот же шаблон с тем же значением: ничего не меняется
	err := tree.ReplaceString("/u/:slot/v2/*path", "/u/:slot/v2/*path", "v3")
	fmt.Println(prefixtree.Code(err) == prefixtree.NotChanged)

	fmt.Println(prefixtree.View(tree))
	// Output:
	// <nil>
	// <nil>
	// <nil>
	// true
	// [^[/u/]:slot[/v2/]*path=v3]
}

// Копия независима от оригинала
func ExampleTree_clone() {
	tree := prefixtree.New()
	fmt.Println(tree.SetString("/a/*path", "a"))

	c := tree.Clone()
	fmt.Println(c.SetString("/b/*path", "b"))
	fmt.Println(tree.DeleteString("/a/*path"))

	fmt.Println(prefixtree.View(tree))
	fmt.Println(prefixtree.View(c))
	// Output:
	// <nil>
	// <nil>
	// <nil>
	// []
	// [^[/][a/]*path=a ^[/][b/]*path=b]
}
