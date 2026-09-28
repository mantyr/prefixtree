package prefixtree

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParams(t *testing.T) {
	Convey("ByName и Map", t, func() {
		ps := Params{{Key: "slot", Value: "abc"}, {Key: "path", Value: "x/y"}}
		value, ok := ps.ByName("slot")
		So(ok, ShouldBeTrue)
		So(value, ShouldEqual, "abc")

		value, ok = ps.ByName("nope")
		So(ok, ShouldBeFalse)
		So(value, ShouldEqual, "")

		So(ps.Map(), ShouldResemble, map[string]string{"slot": "abc", "path": "x/y"})
		So(Params(nil).Map(), ShouldResemble, map[string]string{})
	})

	Convey("Параметры в порядке следования в адресе", t, func() {
		root := newNode()
		So(root.SetString("/u/:slot/api/*path", "uapi"), ShouldBeNil)
		value, err := root.GetString("/u/abc/api/x/y")
		So(err, ShouldBeNil)
		So(value.Path(), ShouldEqual, "/u/abc/api/x/y")
		So(value.Params(), ShouldResemble, Params{{Key: "slot", Value: "abc"}, {Key: "path", Value: "x/y"}})
	})

	Convey("Без параметров Params пустой", t, func() {
		root := newNode()
		So(root.SetString("/api/v1/health", "h"), ShouldBeNil)
		value, err := root.GetString("/api/v1/health")
		So(err, ShouldBeNil)
		So(value.Params(), ShouldBeEmpty)
	})

	Convey("Параметров больше чем буферы поиска и результата, с откатом через границу буфера", t, func() {
		root := newNode()
		So(root.SetString("/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/x", "long"), ShouldBeNil)
		So(root.SetString("/:a/:b/:c/:d/:e/:f/:g/:h/:i/*rest", "fallback"), ShouldBeNil)

		value, err := root.GetString("/1/2/3/4/5/6/7/8/9/10/x")
		So(err, ShouldBeNil)
		So(value.Value(), ShouldEqual, "long")
		So(value.Params(), ShouldResemble, Params{
			{Key: "a", Value: "1"}, {Key: "b", Value: "2"}, {Key: "c", Value: "3"},
			{Key: "d", Value: "4"}, {Key: "e", Value: "5"}, {Key: "f", Value: "6"},
			{Key: "g", Value: "7"}, {Key: "h", Value: "8"}, {Key: "i", Value: "9"},
			{Key: "j", Value: "10"},
		})

		value, err = root.GetString("/1/2/3/4/5/6/7/8/9/10/y")
		So(err, ShouldBeNil)
		So(value.Value(), ShouldEqual, "fallback")
		So(value.Params(), ShouldResemble, Params{
			{Key: "a", Value: "1"}, {Key: "b", Value: "2"}, {Key: "c", Value: "3"},
			{Key: "d", Value: "4"}, {Key: "e", Value: "5"}, {Key: "f", Value: "6"},
			{Key: "g", Value: "7"}, {Key: "h", Value: "8"}, {Key: "i", Value: "9"},
			{Key: "rest", Value: "10/y"},
		})
	})

	Convey("Результаты разных Get не делят параметры", t, func() {
		root := newNode()
		So(root.SetString("/u/:slot/api/*path", "uapi"), ShouldBeNil)
		first, err := root.GetString("/u/one/api/a")
		So(err, ShouldBeNil)
		second, err := root.GetString("/u/two/api/b")
		So(err, ShouldBeNil)
		So(first.Params().Map(), ShouldResemble, map[string]string{"slot": "one", "path": "a"})
		So(second.Params().Map(), ShouldResemble, map[string]string{"slot": "two", "path": "b"})
	})
}

func TestGetAllocs(t *testing.T) {
	Convey("Аллокации Get", t, func() {
		root := newNode()
		for _, path := range []string{"/u/:slot/api/*path", "/api/v1/health"} {
			So(root.SetString(path, path), ShouldBeNil)
		}
		Convey("промах не аллоцирует", func() {
			allocs := testing.AllocsPerRun(100, func() {
				root.GetString("/u/abc/other/x") //nolint:errcheck
			})
			So(allocs, ShouldEqual, 0)
		})
		Convey("найденное значение это одна аллокация", func() {
			for _, path := range []string{"/u/abc/api/x/y", "/api/v1/health"} {
				allocs := testing.AllocsPerRun(100, func() {
					root.GetString(path) //nolint:errcheck
				})
				So(allocs, ShouldEqual, 1)
			}
		})
	})
}
