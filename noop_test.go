package prefixtree

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// testHandler структура, сравнимая по типу, но с несравнимым значением внутри
type testHandler struct {
	fn interface{}
}

func TestSameValue(t *testing.T) {
	Convey("sameValue сравнивает без паники", t, func() {
		f := func() {}
		p := &testHandler{}
		cases := []struct {
			name string
			a, b interface{}
			want bool
		}{
			{"равные строки", "a", "a", true},
			{"разные строки", "a", "b", false},
			{"равные числа", 1, 1, true},
			{"разные типы с одинаковым видом", 1, int64(1), false},
			{"один указатель", p, p, true},
			{"разные указатели", p, &testHandler{}, false},
			{"func всегда разные", f, f, false},
			{"map всегда разные", map[string]int{}, map[string]int{}, false},
			{"slice всегда разные", []int{1}, []int{1}, false},
			{"структура с func внутри", testHandler{fn: f}, testHandler{fn: f}, false},
			{"структура со сравнимым внутри", testHandler{fn: 1}, testHandler{fn: 1}, true},
		}
		for _, c := range cases {
			So(c.name+" "+boolString(sameValue(c.a, c.b)), ShouldEqual, c.name+" "+boolString(c.want))
		}
	})
}

// boolString нужна чтобы в сообщении падения было видно, какой случай упал
func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestReplaceNoop(t *testing.T) {
	Convey("Replace с тем же шаблоном и тем же значением даёт NotChanged и ничего не публикует", t, func() {
		s := NewStorage()
		So(s.Set("/u/:slot/api/*path", "v1"), ShouldBeNil)

		before := s.Tree()
		So(Code(s.Replace("/u/:slot/api/*path", "/u/:slot/api/*path", "v1")), ShouldEqual, NotChanged)
		So(s.Tree(), ShouldPointTo, before)

		So(s.Replace("/u/:slot/api/*path", "/u/:slot/api/*path", "v2"), ShouldBeNil)
		So(s.Tree(), ShouldNotPointTo, before)
		So(valueOf(s.Tree(), "/u/abc/api/x"), ShouldEqual, "v2")
	})

	Convey("Несравнимые значения не роняют Replace и всегда публикуются", t, func() {
		s := NewStorage()
		f := func() {}
		So(s.Set("/a/*x", f), ShouldBeNil)

		before := s.Tree()
		So(func() { So(s.Replace("/a/*x", "/a/*x", f), ShouldBeNil) }, ShouldNotPanic)
		So(s.Tree(), ShouldNotPointTo, before)

		h := testHandler{fn: f}
		So(func() { So(s.Replace("/a/*x", "/a/*x", h), ShouldBeNil) }, ShouldNotPanic)
		So(func() { So(s.Replace("/a/*x", "/a/*x", h), ShouldBeNil) }, ShouldNotPanic)
	})

	Convey("Указатель на обработчик сравнивается по адресу", t, func() {
		s := NewStorage()
		h := &testHandler{}
		So(s.Set("/a/*x", h), ShouldBeNil)
		before := s.Tree()

		So(Code(s.Replace("/a/*x", "/a/*x", h)), ShouldEqual, NotChanged)
		So(s.Tree(), ShouldPointTo, before)

		// другой указатель на такую же структуру это изменение
		So(s.Replace("/a/*x", "/a/*x", &testHandler{}), ShouldBeNil)
		So(s.Tree(), ShouldNotPointTo, before)
	})

	Convey("Replace на самом дереве с тем же значением даёт NotChanged", t, func() {
		tree := New()
		So(tree.SetString("/a/*x", "v"), ShouldBeNil)
		before := View(tree)
		So(Code(tree.ReplaceString("/a/*x", "/a/*x", "v")), ShouldEqual, NotChanged)
		So(Code(tree.Replace([]byte("/a/*x"), []byte("/a/*x"), "v")), ShouldEqual, NotChanged)
		So(View(tree), ShouldResemble, before)
	})

	Convey("NotChanged только для того же шаблона, другой шаблон с тем же значением это изменение", t, func() {
		tree := New()
		So(tree.SetString("/a/*x", "v"), ShouldBeNil)
		So(tree.ReplaceString("/a/*x", "/b/*x", "v"), ShouldBeNil)
		So(valueOf(tree, "/b/1"), ShouldEqual, "v")
	})
}
