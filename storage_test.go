package prefixtree

import (
	"sync"
	"sync/atomic"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// valueOf возвращает найденное значение или nil
func valueOf(tree TreeReadonly, path string) interface{} {
	value, err := tree.GetString(path)
	if err != nil {
		return nil
	}
	return value.Value()
}

// nodesOf возвращает множество всех нод поддерева
func nodesOf(n *treeNode, set map[*treeNode]bool) map[*treeNode]bool {
	set[n] = true
	for _, child := range n.children {
		nodesOf(child, set)
	}
	return set
}

func TestClone(t *testing.T) {
	Convey("Clone совпадает с оригиналом и не ссылается на него", t, func() {
		root := buildTree(deleteRoutes, func(int) bool { return true })
		c := root.clone(nil)

		So(View(c), ShouldResemble, View(root))
		So(checkParents(c, ""), ShouldBeEmpty)

		var d diffs
		for _, path := range deleteProbes {
			d.equal("probe", probe(c, path), probe(root, path))
		}
		original := nodesOf(root, map[*treeNode]bool{})
		for node := range nodesOf(c, map[*treeNode]bool{}) {
			if original[node] || original[node.parent] {
				d.add("clone shares a node with the original: " + node.view())
			}
		}
		So(d.result(), ShouldBeEmpty)
	})

	Convey("Изменения оригинала не видны в копии", t, func() {
		root := buildTree(deleteRoutes, func(int) bool { return true })
		c := root.clone(nil)
		before := snapshotOf(c)
		for _, route := range deleteRoutes {
			So(root.DeleteString(route), ShouldBeNil)
		}
		So(root.SetString("/new/*x", "new"), ShouldBeNil)
		So(snapshotOf(c), ShouldEqual, before)
	})
}

func TestTreeInterface(t *testing.T) {
	Convey("Tree через интерфейс: копия и оригинал меняются независимо", t, func() {
		tree := New()
		So(tree.SetString("/a/*x", "a"), ShouldBeNil)
		So(tree.Set([]byte("/b/*x"), "b"), ShouldBeNil)

		c := tree.Clone()
		So(c.DeleteString("/a/*x"), ShouldBeNil)
		So(c.ReplaceString("/b/*x", "/c/*x", "c"), ShouldBeNil)
		So(tree.Delete([]byte("/b/*x")), ShouldBeNil)
		So(tree.Replace([]byte("/a/*x"), []byte("/d/*x"), "d"), ShouldBeNil)

		So(valueOf(tree, "/a/1"), ShouldBeNil)
		So(valueOf(tree, "/b/1"), ShouldBeNil)
		So(valueOf(tree, "/d/1"), ShouldEqual, "d")

		So(valueOf(c, "/a/1"), ShouldBeNil)
		So(valueOf(c, "/b/1"), ShouldBeNil)
		So(valueOf(c, "/c/1"), ShouldEqual, "c")
		So(valueOf(c, "/d/1"), ShouldBeNil)
	})
}

// foreignTree чужая реализация TreeReadonly
type foreignTree struct{}

func (foreignTree) Get(path []byte) (Value, error)       { return nil, errNotFound }
func (foreignTree) GetString(path string) (Value, error) { return nil, errNotFound }

func TestView(t *testing.T) {
	Convey("View работает по Tree, по копии из Storage и пропускает чужие реализации", t, func() {
		tree := New()
		So(tree.SetString("/u/:slot/api/*path", "api"), ShouldBeNil)
		So(View(tree), ShouldResemble, []string{"^[/u/]:slot[/api/]*path=api"})

		s := NewStorage()
		So(s.Set("/u/:slot/api/*path", "api"), ShouldBeNil)
		So(View(s.Tree()), ShouldResemble, []string{"^[/u/]:slot[/api/]*path=api"})

		So(View(foreignTree{}), ShouldBeNil)
	})
}

func TestStorage(t *testing.T) {
	Convey("Storage", t, func() {
		s := NewStorage()

		Convey("пустое хранилище", func() {
			_, err := s.Tree().GetString("/x")
			So(Code(err), ShouldEqual, PathNotFound)
		})

		Convey("Set, Delete и Replace публикуют новую копию", func() {
			before := s.Tree()
			So(s.Set("/u/:slot/api/*path", "v1"), ShouldBeNil)
			So(s.Tree(), ShouldNotPointTo, before)
			So(valueOf(s.Tree(), "/u/abc/api/x"), ShouldEqual, "v1")

			value, err := s.Tree().Get([]byte("/u/abc/api/x"))
			So(err, ShouldBeNil)
			So(value.Params(), ShouldResemble, Params{{Key: "slot", Value: "abc"}, {Key: "path", Value: "x"}})

			before = s.Tree()
			So(s.Replace("/u/:slot/api/*path", "/u/:slot/v2/*path", "v2"), ShouldBeNil)
			So(s.Tree(), ShouldNotPointTo, before)
			So(valueOf(s.Tree(), "/u/abc/api/x"), ShouldBeNil)
			So(valueOf(s.Tree(), "/u/abc/v2/x"), ShouldEqual, "v2")

			before = s.Tree()
			So(s.Delete("/u/:slot/v2/*path"), ShouldBeNil)
			So(s.Tree(), ShouldNotPointTo, before)
			So(valueOf(s.Tree(), "/u/abc/v2/x"), ShouldBeNil)
		})

		Convey("ошибки возвращаются с кодами и ничего не публикуют", func() {
			So(s.Set("/a/*x", "a"), ShouldBeNil)
			before := s.Tree()
			So(Code(s.Set("/a/*x", "again")), ShouldEqual, PathAlreadyExists)
			So(Code(s.Set("/b/:id/:id", "b")), ShouldEqual, DuplicateParamInPath)
			So(Code(s.Set("/b/*x", nil)), ShouldEqual, EmptyValue)
			So(Code(s.Delete("/nope")), ShouldEqual, PathNotFound)
			So(Code(s.Replace("/nope", "/x/*y", "v")), ShouldEqual, PathNotFound)
			So(Code(s.Replace("/a/*x", "/a?b", "v")), ShouldEqual, InvalidPath)
			So(s.Tree(), ShouldPointTo, before)
			So(valueOf(s.Tree(), "/a/1"), ShouldEqual, "a")
		})

		Convey("старая копия не видит новых изменений", func() {
			So(s.Set("/a/*x", "a"), ShouldBeNil)
			old := s.Tree()
			So(s.Delete("/a/*x"), ShouldBeNil)
			So(s.Set("/b/*x", "b"), ShouldBeNil)

			So(valueOf(old, "/a/1"), ShouldEqual, "a")
			So(valueOf(old, "/b/1"), ShouldBeNil)
			So(valueOf(s.Tree(), "/a/1"), ShouldBeNil)
			So(valueOf(s.Tree(), "/b/1"), ShouldEqual, "b")
		})
	})
}

func TestStorageConcurrent(t *testing.T) {
	Convey("Читатели и писатель параллельно, запускать с -race", t, func() {
		s := NewStorage()
		So(s.Set("/stable/*path", "stable"), ShouldBeNil)
		So(s.Set("/u/:slot/api/*path", "v0"), ShouldBeNil)

		// маршрут /u/:slot/api/*path меняется только через Replace, остальные гоняются по кругу
		var churn []string
		for _, route := range deleteRoutes {
			if route != "/u/:slot/api/*path" {
				churn = append(churn, route)
			}
		}

		var stop atomic.Bool
		var wg sync.WaitGroup
		errs := make(chan string, 16)

		// писатель: маршруты со сплитами и склейками плюс смена значения через Replace
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer stop.Store(true)
			for i := 0; i < 300; i++ {
				for _, route := range churn {
					if err := s.Set(route, route); err != nil {
						errs <- "set " + route + ": " + err.Error()
						return
					}
				}
				for _, route := range churn {
					if err := s.Delete(route); err != nil {
						errs <- "delete " + route + ": " + err.Error()
						return
					}
				}
				version := []string{"v1", "v0"}[i%2]
				if err := s.Replace("/u/:slot/api/*path", "/u/:slot/api/*path", version); err != nil {
					errs <- "replace: " + err.Error()
					return
				}
			}
		}()

		// читатели: берут копию и проверяют инварианты
		for r := 0; r < 8; r++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for !stop.Load() {
					tree := s.Tree()
					if valueOf(tree, "/stable/x") != "stable" {
						errs <- "stable route lost"
						return
					}
					if v := valueOf(tree, "/u/abc/api/x"); v != "v0" && v != "v1" {
						errs <- "api route missing"
						return
					}
					for _, path := range deleteProbes {
						tree.GetString(path) //nolint:errcheck
					}
				}
			}()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			So(err, ShouldBeEmpty)
		}
	})
}
