package prefixtree_test

import (
	"fmt"

	"github.com/mantyr/prefixtree/v2"
)

// Читатели берут опубликованную копию и не ждут писателей
func ExampleStorage() {
	s := prefixtree.NewStorage()
	fmt.Println(s.Set("/u/:slot/api/*path", "v1"))

	// копия, с которой работает, например, текущий запрос
	old := s.Tree()

	fmt.Println(s.Replace("/u/:slot/api/*path", "/u/:slot/api/*path", "v2"))

	value, _ := old.GetString("/u/abc/api/x")
	fmt.Println("old:", value.Value())

	value, _ = s.Tree().GetString("/u/abc/api/x")
	fmt.Println("new:", value.Value())
	// Output:
	// <nil>
	// <nil>
	// old: v1
	// new: v2
}

// Без изменений новая копия не публикуется
func ExampleStorage_replace() {
	type handler struct{ name string }
	h := &handler{name: "api"}

	s := prefixtree.NewStorage()
	fmt.Println(s.Set("/u/:slot/api/*path", h))
	before := s.Tree()

	err := s.Replace("/u/:slot/api/*path", "/u/:slot/api/*path", h)
	fmt.Println(prefixtree.Code(err) == prefixtree.NotChanged)
	fmt.Println(s.Tree() == before)

	fmt.Println(s.Replace("/u/:slot/api/*path", "/u/:slot/api/*path", &handler{name: "api"}))
	fmt.Println(s.Tree() == before)
	// Output:
	// <nil>
	// true
	// true
	// <nil>
	// false
}

// При ошибке мастер не меняется и новая копия не публикуется
func ExampleStorage_set() {
	s := prefixtree.NewStorage()
	fmt.Println(s.Set("/a/*path", "a"))
	before := s.Tree()

	err := s.Set("/a/*path", "again")
	fmt.Println(prefixtree.Code(err) == prefixtree.PathAlreadyExists)
	fmt.Println(s.Tree() == before)
	// Output:
	// <nil>
	// true
	// true
}
