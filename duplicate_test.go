package prefixtree

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDuplicateParamInPath(t *testing.T) {
	Convey("Повторное имя параметра в шаблоне", t, func() {
		root := newNode()
		for _, path := range []string{
			"/a/:id/b/:id",
			"/a/:id/b/:id/c",
			"/a/:id/b/*id",
		} {
			Convey(path, func() {
				So(Code(root.SetString(path, "v")), ShouldEqual, DuplicateParamInPath)
			})
		}

		Convey("дерево не меняется при ошибке", func() {
			So(root.SetString("/a/:id/b/:id", "v"), ShouldNotBeNil)
			So(root.children, ShouldBeEmpty)
		})

		Convey("одинаковые имена в разных шаблонах допустимы", func() {
			So(root.SetString("/a/:id/c", "1"), ShouldBeNil)
			So(root.SetString("/b/:id/c", "2"), ShouldBeNil)
		})

		Convey("внешний параметр не теряется при откате", func() {
			So(root.SetString("/x/:id/b/:name/c", "inner"), ShouldBeNil)
			So(root.SetString("/x/:id/b/*rest", "fallback"), ShouldBeNil)
			value, err := root.GetString("/x/1/b/2/zzz")
			So(err, ShouldBeNil)
			So(value.Value(), ShouldEqual, "fallback")
			So(value.Params(), ShouldResemble, Params{{Key: "id", Value: "1"}, {Key: "rest", Value: "2/zzz"}})
		})
	})
}
