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
					lastNode.Parent = nil
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

func TestGet(t *testing.T) {
	Convey("Проверяем получение значения по адресу", t, func() {
		root := New("", Root)

		node1 := root.Node("/path/")
		node2 := node1.Param("dir").Node("/")
		So(node2.String(), ShouldEqual, "^[/path/]:dir[/]")

		node3 := node2.All("filepath")
		node3.Value = "tree1"
		So(node3.String(), ShouldEqual, "^[/path/]:dir[/]*filepath")

		node4 := node2.Node("123")
		node4.Value = "tree2"
		// данный маршрут перебивает предыдущий если адрес файла 123
		So(node4.String(), ShouldEqual, "^[/path/]:dir[/][123]")

		node5 := node2.Node("file2.zip")
		node5.Value = "tree3"
		// данный маршрут перебивает предыдущий если адрес файла file2.zip
		So(node5.String(), ShouldEqual, "^[/path/]:dir[/][file2.zip]")

		node6 := node1.Node("user_").Param("user")
		node6.Value = "tree4"
		So(node6.String(), ShouldEqual, "^[/path/][user_]:user")

		node7 := root.Node("/id")
		node8 := node7.Node("/").Param("id")
		node8.Value = "tree5"
		So(node8.String(), ShouldEqual, "^[/id][/]:id")

		node9 := node7.Param("id")
		node9.Value = "tree6"
		So(node9.String(), ShouldEqual, "^[/id]:id")

		tests := []struct {
			path          string
			expectedPath  string
			expectedValue string
			expectedError error
		}{
			{
				path:          "/path/123/file1.zip",
				expectedPath:  "^[/path/]:dir[/]*filepath",
				expectedValue: "tree1",
				expectedError: nil,
			},
			{
				path:          "/path/123/123",
				expectedPath:  "^[/path/]:dir[/][123]",
				expectedValue: "tree2",
				expectedError: nil,
			},
			{
				path:          "/path/123/file2.zip",
				expectedPath:  "^[/path/]:dir[/][file2.zip]",
				expectedValue: "tree3",
				expectedError: nil,
			},
			{
				path:          "/path/123/123/123/file.zip",
				expectedPath:  "^[/path/]:dir[/]*filepath",
				expectedValue: "tree1",
				expectedError: nil,
			},
			{
				path:          "/path/test123",
				expectedError: errors.New("not found"),
			},
			{
				path:          "path/123/file.zip",
				expectedError: errors.New("not found"),
			},
			{
				path:          "/id/123",
				expectedPath:  "^[/id][/]:id",
				expectedValue: "tree5",
				expectedError: nil,
			},
		}
		for n, test := range tests {
			Convey(fmt.Sprintf("#%d", n), func() {
				Printf("\n\tpath: %s\n", test.path)

				node, err := root.Get([]byte(test.path))
				if test.expectedError != nil {
					Printf("\tfind: %s\n\t", test.expectedError.Error())
					So(err, ShouldNotBeNil)
					So(
						err.Error(),
						ShouldEqual,
						test.expectedError.Error(),
					)
					So(node, ShouldBeNil)
				} else {
					Printf("\tfind: %s\n\t", test.expectedPath)
					So(err, ShouldBeNil)
					So(node, ShouldNotBeNil)
					So(
						node.String(),
						ShouldResemble,
						test.expectedPath,
					)
					So(
						node.Value,
						ShouldResemble,
						test.expectedValue,
					)
				}
			})
		}
	})
}
