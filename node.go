package prefix

import (
	"errors"
	"fmt"
)

// Node это один элемент параметризованного префиксного дерева
type Node struct {
	// root это ссылка на первый элемент в адресе
	root *Node

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
func (n *Node) Node(
	path string,
	nodeType ...Type,
) *Node {
	var nType Type
	if len(nodeType) > 0 {
		nType = nodeType[0]
	}
	child := &Node{
		Path: []byte(path),
		Type: nType,
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
