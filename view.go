package prefixtree

import (
	"fmt"
)

type view struct {
	result []string
}

// View возвращает текстовое представление дерева
func View(n *Node) []string {
	v := &view{}
	v.view("", n)
	return v.result
}

func (v *view) view(s string, n *Node) {
	s = s + n.String()
	if n.Value != nil {
		v.result = append(
			v.result,
			fmt.Sprintf(
				"%s=%v",
				s,
				n.Value,
			),
		)
	}
	for _, child := range n.Children {
		v.view(s, child)
	}
}
