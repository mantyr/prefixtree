package prefixtree

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetBytes(t *testing.T) {
	Convey("Get по []byte ведёт себя как GetString", t, func() {
		root := newNode()
		So(root.SetString("/u/:slot/api/*path", "uapi"), ShouldBeNil)

		value, err := root.Get([]byte("/u/abc/api/x/y"))
		So(err, ShouldBeNil)
		So(value.Value(), ShouldEqual, "uapi")
		So(value.Path(), ShouldEqual, "/u/abc/api/x/y")
		So(value.Params(), ShouldResemble, Params{{Key: "slot", Value: "abc"}, {Key: "path", Value: "x/y"}})

		_, err = root.Get([]byte("/nope"))
		So(Code(err), ShouldEqual, PathNotFound)
	})
}

// Защитные ветки, до которых публичным API не дойти: декодер отсекает такие шаблоны раньше
func TestGuards(t *testing.T) {
	Convey("Вставка", t, func() {
		static := pathToken{kind: staticToken, title: []byte("/")}
		param := pathToken{kind: paramToken, title: []byte("id")}
		catchAll := pathToken{kind: catchAllToken, title: []byte("path")}

		Convey("после CatchAll ничего нельзя", func() {
			n := &treeNode{pathToken: catchAll}
			_, err := n.insert(static)
			So(Code(err), ShouldEqual, InvalidPath)
		})
		Convey("CatchAll сразу после Param нельзя", func() {
			n := &treeNode{pathToken: param}
			_, err := n.insert(catchAll)
			So(Code(err), ShouldEqual, InvalidPath)
		})
		Convey("Param сразу после Param нельзя", func() {
			n := &treeNode{pathToken: param}
			_, err := n.insert(param)
			So(Code(err), ShouldEqual, InvalidPath)
		})
		Convey("неизвестный тип токена", func() {
			_, err := newNode().insert(pathToken{kind: 99, title: []byte("x")})
			So(Code(err), ShouldEqual, Internal)
		})
	})

	Convey("Поиск", t, func() {
		Convey("CatchAll-нода не спускается дальше", func() {
			n := &treeNode{pathToken: pathToken{kind: catchAllToken}, wildChild: true}
			q := query{path: "/x"}
			So(n.find(&q), ShouldBeNil)
		})
		Convey("адрес закончился раньше чем ветка", func() {
			n := &treeNode{pathToken: pathToken{kind: staticToken}, wildChild: true}
			q := query{path: "/x", offset: 2}
			So(n.find(&q), ShouldBeNil)
		})
	})

	Convey("Текстовое представление неизвестного токена", t, func() {
		token := pathToken{kind: 99, title: []byte("x")}
		So(token.view(), ShouldEqual, "?x")
	})
}
