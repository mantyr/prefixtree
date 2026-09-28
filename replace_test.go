package prefixtree

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// replaceCandidates новые шаблоны: существующие, новые, конфликтующие и под нодой со значением
var replaceCandidates = append(append([]string{}, deleteRoutes...),
	"/u/:id/api/*path",     // другое имя параметра рядом с :slot
	"/u/:slot/api/*other",  // другое имя CatchAll на месте существующего
	"/api/v1/health/*rest", // CatchAll после новой статики
	"/api/v1/health*rest",  // CatchAll прямо под статикой со значением
	"/api/v3/*path",        // новая ветка со сплитом
	"/api/v",               // шаблон кончается посреди ноды
	"/s/*other",            // конфликт CatchAll, если /s/*path остаётся
	"/path/:dir/1234",      // продолжение статической ноды со значением
	"/new/:a/:b/*c",        // совсем новая ветка
	"/id/:id/*rest",        // CatchAll под параметром через статику
	"/a/:x/:x",             // повтор имени
	"/a?b",                 // невалидный символ
	"",                     // пустой
)

// expectedAfterReplace собирает дерево без old и с new, возвращает ошибку Set(new)
func expectedAfterReplace(old, new string) (*treeNode, error) {
	root := buildTree(deleteRoutes, func(i int) bool { return deleteRoutes[i] != old })
	return root, root.SetString(new, "replaced")
}

// snapshotOf возвращает полное состояние дерева для сравнения
func snapshotOf(root *treeNode) string {
	state := fmt.Sprint(View(root))
	for _, path := range deleteProbes {
		state += "\n" + probe(root, path)
	}
	return state
}

func TestReplaceOracle(t *testing.T) {
	Convey("Replace совпадает с деревом, собранным без old и с new, либо не меняет дерево", t, func() {
		var d diffs
		for _, old := range deleteRoutes {
			for _, new := range replaceCandidates {
				root := buildTree(deleteRoutes, func(int) bool { return true })
				before := snapshotOf(root)
				err := root.ReplaceString(old, new, "replaced")
				label := fmt.Sprintf("%q -> %q", old, new)

				if old == new {
					d.equal(label+" err", err, nil)
					tokens, _ := patternTokens([]byte(old))
					d.equal(label+" value", root.lookup(tokens).value, "replaced")
					continue
				}

				expected, expectedErr := expectedAfterReplace(old, new)
				d.equal(label+" code", Code(err), Code(expectedErr))
				if expectedErr != nil {
					d.equal(label+" unchanged", snapshotOf(root), before)
				} else {
					d.equal(label+" view", sortedView(root), sortedView(expected))
					for _, path := range append(deleteProbes, "/u/abc/api/x", "/api/v3/x", "/api/v1/health/x", "/new/1/2/3") {
						d.equal(label+" probe", probe(root, path), probe(expected, path))
					}
				}
				d.addAll(label, checkParents(root, ""))
			}
		}
		So(d.result(), ShouldBeEmpty)
	})
}

func TestReplace(t *testing.T) {
	Convey("Replace", t, func() {
		root := newNode()
		So(root.SetString("/u/:slot/oauth/*path", "oauth"), ShouldBeNil)
		So(root.SetString("/u/:slot/api/*path", "api"), ShouldBeNil)

		Convey("тот же шаблон меняет значение и сохраняет место среди соседей", func() {
			So(root.SetString("/id:id", "first"), ShouldBeNil)
			So(root.SetString("/id:id2", "second"), ShouldBeNil)
			So(root.ReplaceString("/id:id", "/id:id", "first-v2"), ShouldBeNil)
			value, err := root.GetString("/id9")
			So(err, ShouldBeNil)
			So(value.Value(), ShouldEqual, "first-v2")
		})

		Convey("переименование параметра рядом с тем же именем в соседнем маршруте", func() {
			So(root.ReplaceString("/u/:slot/api/*path", "/u/:id/api/*path", "api-v2"), ShouldBeNil)
			value, err := root.GetString("/u/abc/api/x")
			So(err, ShouldBeNil)
			So(value.Value(), ShouldEqual, "api-v2")
			So(value.Params(), ShouldResemble, Params{{Key: "id", Value: "abc"}, {Key: "path", Value: "x"}})
			value, err = root.GetString("/u/abc/oauth/x")
			So(err, ShouldBeNil)
			So(value.Params(), ShouldResemble, Params{{Key: "slot", Value: "abc"}, {Key: "path", Value: "x"}})
		})

		Convey("CatchAll на месте старого с другим именем", func() {
			So(root.ReplaceString("/u/:slot/api/*path", "/u/:slot/api/*rest", "api-v2"), ShouldBeNil)
			value, err := root.GetString("/u/abc/api/x")
			So(err, ShouldBeNil)
			So(value.Params(), ShouldResemble, Params{{Key: "slot", Value: "abc"}, {Key: "rest", Value: "x"}})
		})

		Convey("CatchAll под статической нодой со значением", func() {
			So(root.SetString("/a/", "a"), ShouldBeNil)
			So(root.SetString("/b/*x", "b"), ShouldBeNil)
			So(root.ReplaceString("/b/*x", "/a/*x", "ax"), ShouldBeNil)
			value, err := root.GetString("/a/1")
			So(err, ShouldBeNil)
			So(value.Value(), ShouldEqual, "ax")
		})

		Convey("ошибки не меняют дерево", func() {
			before := snapshotOf(root)
			So(Code(root.ReplaceString("/nope", "/x/*y", "v")), ShouldEqual, PathNotFound)
			So(Code(root.ReplaceString("/u/:slot/api/*path", "/u/:slot/oauth/*path", "v")), ShouldEqual, PathAlreadyExists)
			So(Code(root.ReplaceString("/u/:slot/api/*path", "/u/:slot/oauth/*other", "v")), ShouldEqual, PathAlreadyExists)
			So(Code(root.ReplaceString("/u/:slot/api/*path", "/a/:x/:x", "v")), ShouldEqual, DuplicateParamInPath)
			So(Code(root.ReplaceString("/u/:slot/api/*path", "/a?b", "v")), ShouldEqual, InvalidPath)
			So(Code(root.ReplaceString("/u/:slot/api/*path", "", "v")), ShouldEqual, InvalidPath)
			So(Code(root.ReplaceString("", "/x", "v")), ShouldEqual, InvalidPath)
			So(Code(root.ReplaceString("/u/:slot/api/*path", "/x", nil)), ShouldEqual, EmptyValue)
			So(snapshotOf(root), ShouldEqual, before)
		})

		Convey("Replace по []byte копирует новый шаблон", func() {
			buf := []byte("/u/:slot/v2/*path")
			So(root.Replace([]byte("/u/:slot/api/*path"), buf, "v2"), ShouldBeNil)
			copy(buf, "/x/:zzzz/xx/*xxxx")
			value, err := root.GetString("/u/abc/v2/x")
			So(err, ShouldBeNil)
			So(value.Value(), ShouldEqual, "v2")
		})
	})

	Convey("mustInsert падает на ошибке", t, func() {
		So(func() { mustInsert(errNotFound) }, ShouldPanic)
		So(func() { mustInsert(nil) }, ShouldNotPanic)
	})
}

func TestStorageReplace(t *testing.T) {
	Convey("Storage.Replace", t, func() {
		s := NewStorage()
		So(s.Set("/u/:slot/api/*path", "v1"), ShouldBeNil)

		before := s.Tree()
		So(s.Replace("/u/:slot/api/*path", "/u/:slot/api/*path", "v2"), ShouldBeNil)
		So(valueOf(before, "/u/abc/api/x"), ShouldEqual, "v1")
		So(valueOf(s.Tree(), "/u/abc/api/x"), ShouldEqual, "v2")

		So(s.Replace("/u/:slot/api/*path", "/u/:slot/v3/*path", "v3"), ShouldBeNil)
		So(valueOf(s.Tree(), "/u/abc/api/x"), ShouldBeNil)
		So(valueOf(s.Tree(), "/u/abc/v3/x"), ShouldEqual, "v3")

		published := s.Tree()
		So(Code(s.Replace("/nope", "/x/*y", "v")), ShouldEqual, PathNotFound)
		So(s.Tree(), ShouldPointTo, published)
	})
}
