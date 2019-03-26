package prefixtree

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// testSetValue проверяет вставку в дерево
func testSetValue(
	root *Node,
	path string,
	value string,
	expectedError error,
) {
	err := root.SetString(path, value)
	if expectedError != nil {
		Convey(fmt.Sprintf("\t%s - (error)", path), func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, expectedError.Error())
		})
	} else {
		Convey(fmt.Sprintf("\t%s", path), func() {
			So(err, ShouldBeNil)
		})
	}
}

// testGetValue проверяет получение данных из дерева
func testGetValue(
	root *Node,
	path string,
	expectedError error,
	expectedValue interface{},
	expectedParams map[string]string,
) {
	value, err := root.GetString(path)
	if expectedError != nil {
		Convey(fmt.Sprintf("%s - (error)", path), func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, expectedError.Error())
		})
	} else {
		Convey(path, func() {
			So(err, ShouldBeNil)
			So(value, ShouldNotBeNil)
			So(value.Value, ShouldNotBeNil)
			So(
				value.Value,
				ShouldResemble,
				expectedValue,
			)
			So(value.Params, ShouldNotBeNil)
			So(
				value.Params,
				ShouldResemble,
				expectedParams,
			)
		})
	}
}

func TestNode(t *testing.T) {
	Convey("Проверяем вставку в дерево", t, func() {
		root := New()
		So(root, ShouldNotBeNil)
		So(root.Type, ShouldEqual, Root)

		testSetValue(root, "/path/:dir/123", "value1", nil)
		testSetValue(root, "/path/:dir/*filepath", "value2", nil)
		testSetValue(root, "/path/user_:user", "value3", nil)
		testSetValue(root, "/id/:id", "value4", nil)
		testSetValue(root, "/id:id", "value5", nil)
		testSetValue(root, ":id/:name/123", "value6", nil)
		testSetValue(root, "/id:id", "value7", errors.New("path already in use"))
		testSetValue(root, "/id:id2", "value8", nil)

		Convey("Проверяем содержимое дерева", func() {
			So(
				View(root),
				ShouldResemble,
				[]string{
					"^[/][path/]:dir[/][123]=value1", // а куда потерялся /123 ?
					"^[/][path/]:dir[/]*filepath=value2",
					"^[/][path/][user_]:user=value3",
					"^[/][id][/]:id=value4",
					"^[/][id]:id=value5",
					"^[/][id]:id2=value8",
					"^:id[/]:name[/123]=value6",
				},
			)
		})
		Convey("Проверяем получение данных из дерева", func() {
			Convey("Нет данных", func() {
				testGetValue(root, "/", errors.New("not found"), nil, nil)
				testGetValue(root, "/path", errors.New("not found"), nil, nil)
				testGetValue(root, "/path/", errors.New("not found"), nil, nil)
				testGetValue(root, ":id", errors.New("not found"), nil, nil)
			})
			Convey("Данные найдены", func() {
				testGetValue(
					root,
					"/path/123/123",
					nil,
					"value1",
					map[string]string{
						"dir": "123",
					},
				)
				testGetValue(
					root,
					"/path/test/123",
					nil,
					"value1",
					map[string]string{
						"dir": "test",
					},
				)
				testGetValue(
					root,
					"/path/path/123",
					nil,
					"value1",
					map[string]string{
						"dir": "path",
					},
				)
				testGetValue(
					root,
					"/path/:id/123",
					nil,
					"value1",
					map[string]string{
						"dir": ":id",
					},
				)
				testGetValue(
					root,
					"/path/:id/123/dir/file.zip",
					nil,
					"value2",
					map[string]string{
						"dir":      ":id",
						"filepath": "123/dir/file.zip",
					},
				)
				testGetValue(
					root,
					"/path/user_123",
					nil,
					"value3",
					map[string]string{
						"user": "123",
					},
				)
				testGetValue(
					root,
					"/path/user_123/",
					errors.New("not found"),
					nil,
					nil,
				)
				testGetValue(
					root,
					"/id/123",
					nil,
					"value4",
					map[string]string{
						"id": "123",
					},
				)
				testGetValue(
					root,
					"/id123",
					nil,
					"value5",
					map[string]string{
						"id": "123",
					},
				)
			})
		})
	})
}
