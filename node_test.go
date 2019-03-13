package prefix

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInsert(t *testing.T) {
	Convey("Проверяем вставку адреса", t, func() {
		tests := []struct {
			path             string
			expected         *Node
			expectedLastNode *Node
			expectedError    error
		}{
			{
				path:     "/path/:dir/*filepath",
				expected: New("", Root).Node("/path/").Param("dir").Node("/").All("filepath"),
				expectedLastNode: &Node{
					Path: []byte("filepath"),
					Type: CatchAll,
				},
				expectedError: nil,
			},
			{
				path:     "/path/user_:user",
				expected: New("", Root).Node("/path/user_").Param("user"),
				expectedLastNode: &Node{
					Path: []byte("user"),
					Type: Param,
				},
				expectedError: nil,
			},
			{
				path:     "/id/:id",
				expected: New("", Root).Node("/id/").Param("id"),
				expectedLastNode: &Node{
					Path: []byte("id"),
					Type: Param,
				},
				expectedError: nil,
			},
			{
				path:     "/id:id",
				expected: New("", Root).Node("/id").Param("id"),
				expectedLastNode: &Node{
					Path: []byte("id"),
					Type: Param,
				},
				expectedError: nil,
			},
			{
				path:          "/id:id:name",
				expectedError: errors.New("expected '/' or end path but actual ':' or '*'"),
			},
			{
				path:     ":id/:name/123",
				expected: New("", Root).Param("id").Node("/").Param("name").Node("/123"),
				expectedLastNode: &Node{
					Path: []byte("/123"),
				},
				expectedError: nil,
			},
			{
				path:          ":/",
				expectedError: errors.New("empty param name"),
			},
			{
				path:          ":",
				expectedError: errors.New("empty param name"),
			},
			{
				path:          "::",
				expectedError: errors.New("empty param name"),
			},
			{
				path:          ":id:name/123",
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
				root := &Node{
					Path: []byte{},
					Type: Root,
				}
				lastNode, err := root.Insert([]byte(test.path))
				So(root.root, ShouldBeNil)

				if test.expectedError != nil {
					So(err, ShouldNotBeNil)
					So(
						err.Error(),
						ShouldEqual,
						test.expectedError.Error(),
					)
					So(lastNode, ShouldBeNil)
					So(
						root,
						ShouldResemble,
						&Node{
							Path: []byte{},
							Type: Root,
						},
					)
				} else {
					So(err, ShouldBeNil)
					So(lastNode, ShouldNotBeNil)
					So(
						string(lastNode.Path),
						ShouldEqual,
						string(test.expectedLastNode.Path),
					)
					So(
						root,
						ShouldResemble,
						test.expected.Root(),
					)
					lastNode.root = nil
					So(
						lastNode,
						ShouldResemble,
						test.expectedLastNode,
					)
				}
			})
		}
	})
}
