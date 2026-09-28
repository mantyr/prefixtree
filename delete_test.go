package prefixtree

import (
	"fmt"
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// deleteRoutes маршруты со сплитами статики, соседними параметрами и CatchAll рядом со статикой
var deleteRoutes = []string{
	"/s/*path",
	"/u/:slot/oauth/*path",
	"/u/:slot/api/*path",
	"/api/v1/*path",
	"/api/v2/*path",
	"/api/v1/health",
	"/path/:dir/123",
	"/path/user_:user",
	"/id/:id",
	"/id:id",
	"/id:id2",
	":id/:name/123",
}

// deleteProbes адреса, по которым сравниваются деревья
var deleteProbes = []string{
	"/s/a/b", "/u/abc/oauth/x", "/u/abc/api/x", "/api/v1/x", "/api/v2/y",
	"/api/v1/health", "/path/5/123", "/path/user_7", "/id/9", "/id9",
	"x/y/123", "/u/", "/api/v", "/", "/id", "",
}

// probe возвращает результат Get одной строкой для сравнения
func probe(root *treeNode, path string) string {
	value, err := root.GetString(path)
	if err != nil {
		return fmt.Sprintf("%s -> code %d", path, Code(err))
	}
	return fmt.Sprintf("%s -> %v %v", path, value.Value(), value.Params())
}

// sortedView возвращает View без учёта порядка статических соседей, он не влияет на поиск
func sortedView(root *treeNode) []string {
	items := View(root)
	sort.Strings(items)
	return items
}

// buildTree собирает дерево из маршрутов, для которых keep вернул true, падает если маршрут не встал
func buildTree(routes []string, keep func(i int) bool) *treeNode {
	root := newNode()
	for i, route := range routes {
		if !keep(i) {
			continue
		}
		if err := root.SetString(route, route); err != nil {
			panic(fmt.Sprintf("buildTree %q: %v", route, err))
		}
	}
	return root
}

func TestDeleteEquivalence(t *testing.T) {
	Convey("Дерево после удаления совпадает с деревом, собранным без удалённых маршрутов", t, func() {
		var d diffs
		for mask := 0; mask < 1<<len(deleteRoutes); mask++ {
			deleted := func(i int) bool { return mask&(1<<i) != 0 }
			expected := buildTree(deleteRoutes, func(i int) bool { return !deleted(i) })

			for _, reverse := range []bool{false, true} {
				root := buildTree(deleteRoutes, func(int) bool { return true })
				for k := range deleteRoutes {
					i := k
					if reverse {
						i = len(deleteRoutes) - 1 - k
					}
					if deleted(i) {
						d.equal(fmt.Sprintf("delete %q", deleteRoutes[i]), root.DeleteString(deleteRoutes[i]), nil)
					}
				}
				label := fmt.Sprintf("mask=%b reverse=%v", mask, reverse)
				d.equal(label+" view", sortedView(root), sortedView(expected))
				for _, path := range deleteProbes {
					d.equal(label+" probe", probe(root, path), probe(expected, path))
				}
				d.addAll(label, checkParents(root, ""))
			}
		}
		So(d.result(), ShouldBeEmpty)
	})
}

func TestDelete(t *testing.T) {
	Convey("Delete", t, func() {
		root := newNode()
		for _, route := range []string{"/u/:slot/oauth/*path", "/u/:slot/api/*path", "/api/v1/*path"} {
			So(root.SetString(route, route), ShouldBeNil)
		}

		Convey("удаляет только указанный маршрут", func() {
			So(root.DeleteString("/u/:slot/api/*path"), ShouldBeNil)
			_, err := root.GetString("/u/abc/api/x")
			So(Code(err), ShouldEqual, PathNotFound)
			value, err := root.GetString("/u/abc/oauth/x")
			So(err, ShouldBeNil)
			So(value.Value(), ShouldEqual, "/u/:slot/oauth/*path")
		})

		Convey("повторное удаление даёт PathNotFound", func() {
			So(root.DeleteString("/api/v1/*path"), ShouldBeNil)
			So(Code(root.DeleteString("/api/v1/*path")), ShouldEqual, PathNotFound)
		})

		Convey("после удаления маршрут можно добавить снова", func() {
			So(root.DeleteString("/api/v1/*path"), ShouldBeNil)
			So(root.SetString("/api/v1/*path", "again"), ShouldBeNil)
			value, err := root.GetString("/api/v1/x")
			So(err, ShouldBeNil)
			So(value.Value(), ShouldEqual, "again")
		})

		Convey("Delete по []byte", func() {
			So(root.Delete([]byte("/api/v1/*path")), ShouldBeNil)
			_, err := root.GetString("/api/v1/x")
			So(Code(err), ShouldEqual, PathNotFound)
		})

		Convey("удалить можно только существующий шаблон", func() {
			for _, path := range []string{
				"/nope",                // нет такого
				"/u/:id/api/*path",     // другое имя параметра
				"/u/:slot/api/*other",  // другое имя CatchAll
				"/u/",                  // промежуточная нода без значения
				"/u",                   // шаблон кончается посреди ноды
				"/u/:slot/api/*path/x", // длиннее существующего (и невалиден)
			} {
				err := root.DeleteString(path)
				So(Code(err), ShouldBeIn, []interface{}{PathNotFound, InvalidPath})
			}
			So(Code(root.DeleteString("/nope")), ShouldEqual, PathNotFound)
			So(Code(root.DeleteString("/u/:id/api/*path")), ShouldEqual, PathNotFound)
			So(Code(root.DeleteString("/u")), ShouldEqual, PathNotFound)
		})

		Convey("невалидный шаблон даёт InvalidPath", func() {
			So(Code(root.DeleteString("")), ShouldEqual, InvalidPath)
			So(Code(root.DeleteString("/a?b")), ShouldEqual, InvalidPath)
			So(Code(root.DeleteString("/a/*x/b")), ShouldEqual, InvalidPath)
		})

		Convey("удаление всех маршрутов оставляет пустое дерево", func() {
			for _, route := range []string{"/u/:slot/oauth/*path", "/u/:slot/api/*path", "/api/v1/*path"} {
				So(root.DeleteString(route), ShouldBeNil)
			}
			So(View(root), ShouldBeEmpty)
			So(root.children, ShouldBeEmpty)
			So(root.wildChild, ShouldBeFalse)
		})
	})
}
