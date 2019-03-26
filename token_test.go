package prefixtree

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNodes(t *testing.T) {
	Convey("Проверяем разворачивание path в Tokens", t, func() {
		tests := []struct {
			path          string
			expectedPath  string
			expectedError error
		}{
			{
				path:          "/path/:dir/*filepath",
				expectedPath:  "[/path/]:dir[/]*filepath",
				expectedError: nil,
			},
			{
				path:          "/path/:dir/*filepath?",
				expectedError: errors.New(`unexpected char "?"`),
			},
			{
				path:          "/path/:dir/*filepath     123",
				expectedError: errors.New(`unexpected char " "`),
			},
			{
				path:          "/path/user_:user",
				expectedPath:  "[/path/user_]:user",
				expectedError: nil,
			},
			{
				path:          "/path/user_:user/id",
				expectedPath:  "[/path/user_]:user[/id]",
				expectedError: nil,
			},
			{
				path:          "/path/user_:user/id/:id",
				expectedPath:  "[/path/user_]:user[/id/]:id",
				expectedError: nil,
			},
			{
				path:          "/id:id:name",
				expectedError: errors.New("expected Static token"),
			},
			{
				path:          "/id:",
				expectedError: errors.New("empty token value"),
			},
			{
				path:          "/id:/",
				expectedError: errors.New("empty token value"),
			},
			{
				path:          "/id/*",
				expectedError: errors.New("empty token value"),
			},
		}
		for n, test := range tests {
			Printf("\n\t#%d - %s ", n+1, test.path)
			if test.expectedError != nil {
				Printf("(error) ")
			}
			d := NewDecoder([]byte(test.path))
			So(d, ShouldNotBeNil)

			tokens, err := d.Tokens()
			if test.expectedError != nil {
				So(err, ShouldNotBeNil)
				So(
					err.Error(),
					ShouldEqual,
					test.expectedError.Error(),
				)
			} else {
				So(err, ShouldBeNil)
				So(tokens, ShouldNotBeNil)
				So(
					tokens.String(),
					ShouldResemble,
					test.expectedPath,
				)
			}
		}
	})
}
