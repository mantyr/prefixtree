package prefixtree

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// tokenPattern возвращает кусок шаблона, который даёт нода
func tokenPattern(n *treeNode) string {
	switch n.kind {
	case staticToken:
		return string(n.title)
	case paramToken:
		return ":" + string(n.title)
	case catchAllToken:
		return "*" + string(n.title)
	}
	return ""
}

// patternFromLeaf собирает шаблон адреса, поднимаясь от ноды к корню по Parent
func patternFromLeaf(n *treeNode) string {
	var pattern string
	for ; n != nil; n = n.parent {
		pattern = tokenPattern(n) + pattern
	}
	return pattern
}

// checkParents спускается по children и возвращает ноды, у которых шаблон по parent не сходится с шаблоном сверху
func checkParents(n *treeNode, prefix string) []string {
	var result []string
	for _, child := range n.children {
		expected := prefix + tokenPattern(child)
		if got := patternFromLeaf(child); got != expected {
			result = append(result, fmt.Sprintf("parent chain %q, want %q", got, expected))
		}
		result = append(result, checkParents(child, expected)...)
	}
	return result
}

func TestParentAfterSplit(t *testing.T) {
	Convey("Parent после деления статической ноды", t, func() {
		root := newNode()
		patterns := []string{
			"/api/v1/*path",
			"/api/v2/*path",
			"/u/:slot/oauth/*path",
			"/u/:slot/api/*path",
			"/s/*path",
		}
		for _, pattern := range patterns {
			So(root.SetString(pattern, pattern), ShouldBeNil)
		}

		Convey("у каждого потомка корректный Parent", func() {
			So(checkParents(root, ""), ShouldBeEmpty)
		})

		Convey("шаблон восстанавливается от листа к корню", func() {
			for _, path := range []string{
				"/api/v1/x",
				"/api/v2/x",
				"/u/abc/oauth/x",
				"/u/abc/api/x",
				"/s/x",
			} {
				value, err := root.GetString(path)
				So(err, ShouldBeNil)
				So(patternFromLeaf(value.(*match).node), ShouldEqual, value.Value())
			}
		})
	})
}
