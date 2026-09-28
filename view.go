package prefixtree

import (
	"fmt"
)

type view struct {
	result []string
}

// View возвращает текстовое представление дерева, для чужих реализаций TreeReadonly возвращает nil
func View(t TreeReadonly) []string {
	n, ok := t.(*treeNode)
	if !ok {
		return nil
	}
	v := &view{}
	v.view("", n)
	return v.result
}

func (v *view) view(s string, n *treeNode) {
	s = s + n.view()
	if n.value != nil {
		v.result = append(
			v.result,
			fmt.Sprintf(
				"%s=%v",
				s,
				n.value,
			),
		)
	}
	for _, child := range n.children {
		v.view(s, child)
	}
}
