package prefixtree

import (
	"errors"
	"testing"

	"github.com/mantyr/codes"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCodes(t *testing.T) {
	Convey("Коды пакета не пересекаются со стандартными", t, func() {
		for _, code := range []codes.Code{PathNotFound, PathAlreadyExists, InvalidPath, EmptyValue, DuplicateParamInPath, NotChanged} {
			So(code, ShouldBeGreaterThan, codes.CustomCodes)
		}
	})

	Convey("Set возвращает коды", t, func() {
		root := newNode()
		So(Code(root.SetString("/u/:slot/api/*path", "uapi")), ShouldEqual, OK)
		So(Code(root.SetString("/u/:slot/api/*path", "again")), ShouldEqual, PathAlreadyExists)
		So(Code(root.SetString("/u/:slot/api/*other", "again")), ShouldEqual, PathAlreadyExists)
		So(Code(root.SetString("/s/*path", nil)), ShouldEqual, EmptyValue)

		for _, path := range []string{
			"",        // empty path
			"/a?b",    // unexpected char
			"/a/*x/b", // expected EOF
			"/a/:a:b", // expected staticToken token
			"/a/:",    // empty token value
			"/a/*",    // empty token value
		} {
			Convey(path, func() {
				So(Code(root.SetString(path, "v")), ShouldEqual, InvalidPath)
			})
		}
	})

	Convey("Get возвращает коды", t, func() {
		root := newNode()
		So(root.SetString("/u/:slot/api/*path", "uapi"), ShouldBeNil)

		_, err := root.GetString("/u/abc/api/x")
		So(Code(err), ShouldEqual, OK)

		for _, path := range []string{"/", "/u/abc/oauth/x", "/u/abc/api/", "/nope"} {
			Convey(path, func() {
				_, err := root.GetString(path)
				So(Code(err), ShouldEqual, PathNotFound)
			})
		}
	})

	Convey("Чужая ошибка — Unknown", t, func() {
		So(Code(errors.New("foreign")), ShouldEqual, Unknown)
	})
}
