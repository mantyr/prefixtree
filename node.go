package prefix

import (
	"bytes"
	"errors"
	"fmt"
)

// Node это один элемент параметризованного префиксного дерева
type Node struct {
	// root это ссылка на первый элемент в адресе
	root *Node

	// Parent это ссылка на вышестоящий элемент в адресе
	Parent *Node

	// Path это адрес
	Path []byte

	// Indices это индекс
	// Для проверки, существует ли дочерний элемент со следующим байтом пути
	Indices string

	// Children это вложенные ноды
	Children []*Node

	// Type это тип ноды
	Type Type

	// WildChild это флаг указывающий на наличие потомков
	WildChild bool

	// Value это хранимое в роутинге значение
	Value interface{}
}

// New возвращает новую ноду заданного типа
func New(
	path string,
	nodeType ...Type,
) *Node {
	var nType Type
	if len(nodeType) > 0 {
		nType = nodeType[0]
	}
	return &Node{
		Path: []byte(path),
		Type: nType,
	}
}

// Node добавляет потомка
// осторожно, эта функция только для тестов (внутри panic)
func (n *Node) Node(
	path string,
	nodeType ...Type,
) *Node {
	var nType Type
	if len(nodeType) > 0 {
		nType = nodeType[0]
	}
	switch {
	case n.Type == CatchAll:
		panic("CatchAll нельзя продолжать")
	case path == "":
		panic("empty path")
	case n.Type == Param && path[0] != '/':
		panic("expected '/' or end")
	case n.Type == Param && nType == Param:
		panic("double param")
	}

	child := &Node{
		Path:   []byte(path),
		Type:   nType,
		Parent: n,
	}
	if n.root != nil {
		child.root = n.root
	} else {
		child.root = n
	}
	n.Children = append(n.Children, child)
	return child
}

// Param добавляет параметризированную ноду в потомки
func (n *Node) Param(path string) *Node {
	return n.Node(path, Param)
}

// All добавляет ноду которая забирает в себя весь оставшийся адрес
func (n *Node) All(path string) *Node {
	return n.Node(path, CatchAll)
}

// Root возвращает корневой элемент либо текущий если он первый в цепочке
func (n *Node) Root() *Node {
	root := n.root
	if root == nil {
		root = n
	}
	return root
}

// Insert добавляет цепочку потомков на основе адреса
func (n *Node) Insert(
	path []byte,
) (
	lastNode *Node,
	err error,
) {
	if len(path) == 0 {
		return nil, errors.New("empty path")
	}
	node, err := First(path)
	if err != nil {
		return nil, err
	}
	prefix := len(node.Path)
	switch node.Type {
	case Param, CatchAll:
		prefix++
	}
	save := n.Children
	n.Children = append(n.Children, node)
	if n.root != nil {
		node.root = n.root
	} else {
		node.root = n
	}
	node.Parent = n
	switch {
	case prefix > len(path):
		return nil, fmt.Errorf(
			"unexpected error: long node.Path, actual %s",
			string(node.Path),
		)
	case prefix == len(path):
		return node, nil
	}
	lastNode, err = node.Insert(path[prefix:])
	if err != nil {
		n.Children = save
	}
	return lastNode, err
}

// Childrens возвращает потомков определённого типа
func (n *Node) Childrens(nType Type) []*Node {
	if len(n.Children) == 0 {
		return []*Node{}
	}
	result := make([]*Node, len(n.Children))
	var i int
	for _, node := range n.Children {
		if node.Type == nType {
			result[i] = node
			i++
		}
	}
	return result[:i]
}

// String возвращает адрес ноды
// todo: функция не безопасна для замыкания - это надо поправить
func (n *Node) String() string {
	var title string
	switch n.Type {
	case Root:
		title = "^" + string(n.Path)
	case Param:
		title = ":" + string(n.Path)
	case CatchAll:
		title = "*" + string(n.Path)
	case Static:
		title = "[" + string(n.Path) + "]"
	default:
		title = "???" + string(n.Path)
	}
	if n.Parent != nil {
		return n.Parent.String() + title
	}
	return title
}

// Get возвращает первое полное вхождение адреса
// Static ноды имеют приоритет перед Param
// Param ноды имеют приоритет над CatchAll
func (n *Node) Get(path []byte) (v *Node, err error) {
	//	fmt.Printf("\n%s -> %s\n", n.String(), string(path))
	prefix := len(path)
	for _, node := range n.Childrens(Static) {
		switch {
		case prefix < len(node.Path):
			continue
		case bytes.Compare(path[:len(node.Path)], node.Path) != 0:
			continue
		case prefix == len(node.Path):
			return node, nil
		}
		// ищем первое полное вхождение адреса
		v, err := node.Get(path[len(node.Path):])
		if err == nil {
			return v, nil
		}
	}
	// в Static не нашли, начинаем искать по параметрам
	// случай когда весь path это значение параметра
	slash := bytes.IndexAny(path, "/")
	if slash < 0 {
		for _, node := range n.Childrens(Param) {
			if node.Value != nil {
				return node, nil
			}
		}
		for _, node := range n.Childrens(CatchAll) {
			return node, nil
		}
		return nil, errors.New("not found")
	}
	for _, node := range n.Childrens(Param) {
		v, err := node.Get(path[slash:])
		if err == nil {
			return v, nil
		}
	}
	for _, node := range n.Childrens(CatchAll) {
		return node, nil
	}
	return nil, errors.New("not found")
}
