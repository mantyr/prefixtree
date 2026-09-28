package prefixtree

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSetBytesNotAliased(t *testing.T) {
	Convey("Set не держит память вызывающего", t, func() {
		root := newNode()
		buf := []byte("/u/:slot/api/*path")
		So(root.Set(buf, "uapi"), ShouldBeNil)
		copy(buf, "/x/:zzzz/xxx/*xxxx")

		value, err := root.GetString("/u/abc/api/x")
		So(err, ShouldBeNil)
		So(value.Params(), ShouldResemble, Params{{Key: "slot", Value: "abc"}, {Key: "path", Value: "x"}})
		So(View(root), ShouldResemble, []string{"^[/u/]:slot[/api/]*path=uapi"})
	})
}
