package prefix

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFirst(t *testing.T) {
	Convey("Проверяем определение первой ноды в адресе", t, func() {
		tests := []struct {
			path          string
			expected      *Node
			expectedError error
		}{
			{
				path: "/path/:dir/*filepath",
				expected: &Node{
					Path: []byte("/path/"),
				},
				expectedError: nil,
			},
			{
				path: "/path/user_:user",
				expected: &Node{
					Path: []byte("/path/user_"),
				},
				expectedError: nil,
			},
			{
				path: "/id/:id",
				expected: &Node{
					Path: []byte("/id/"),
				},
				expectedError: nil,
			},
			{
				path: "/id:id",
				expected: &Node{
					Path: []byte("/id"),
				},
				expectedError: nil,
			},
			{
				path: "/id:id:name",
				expected: &Node{
					Path: []byte("/id"),
				},
				expectedError: nil,
			},
			{
				path: ":id/:name/123",
				expected: &Node{
					Path: []byte("id"),
					Type: Param,
				},
				expectedError: nil,
			},
			{
				path:          ":/",
				expected:      nil,
				expectedError: errors.New("empty param name"),
			},
			{
				path:          ":",
				expected:      nil,
				expectedError: errors.New("empty param name"),
			},
			{
				path:          "::",
				expected:      nil,
				expectedError: errors.New("empty param name"),
			},
			{
				path:          ":id:name/123",
				expected:      nil,
				expectedError: errors.New("expected '/' or end path but actual ':' or '*'"),
			},
		}
		for _, test := range tests {
			title := test.path
			if test.expectedError != nil {
				title = fmt.Sprintf(
					"%s - %s",
					test.path,
					test.expectedError.Error(),
				)
			}
			Convey(title, func() {
				result, err := First([]byte(test.path))
				if test.expectedError != nil {
					So(err, ShouldNotBeNil)
					So(
						err.Error(),
						ShouldEqual,
						test.expectedError.Error(),
					)
					So(result, ShouldBeNil)
				} else {
					So(err, ShouldBeNil)
					So(result, ShouldNotBeNil)
					So(
						string(result.Path),
						ShouldEqual,
						string(test.expected.Path),
					)
					So(
						result,
						ShouldResemble,
						test.expected,
					)
				}
			})
		}
	})
}
