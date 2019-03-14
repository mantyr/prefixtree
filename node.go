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

// Len возвращает длину исходного шаблона
func (n *Node) Len() int {
	switch n.Type {
	case Param, CatchAll:
		return len(n.Path) + 1
	}
	return len(n.Path)
}

// Insert добавляет цепочку потомков на основе адреса и возвращает последний элемент
func (n *Node) Insert(
	path []byte,
) (
	lastNode *Node,
	err error,
) {
	save := n.Children

	lastNode, err = Nodes(path)
	if err != nil {
		return nil, err
	}
	root := lastNode.Root()
	if root == nil {
		return nil, errors.New("empty root")
	}
	node, err := n.Glue(root)
	if err != nil {
		n.Children = save
		return nil, err
	}
	if node != nil {
		// в дереве нашлось точно такая же ветка
		return node, nil
	}
	return lastNode, nil
}

// Glue склеивает дерево с цепочкой,
// возвращает последнюю ноду если произошло полное совпадение
func (n *Node) Glue(
	node *Node,
) (
	lastNode *Node,
	err error,
) {
	switch {
	case n.Type == CatchAll:
		return nil, errors.New("last node CatchAll")
	case node.Type == CatchAll && node.WildChild:
		return nil, errors.New("CatchAll excludes Children")
	}
	switch node.Type {
	case Static:
		max, child := n.ChildIdentic(node.Path)
		if child != nil {
			switch {
			case bytes.Equal(child.Path, node.Path):
				if !node.WildChild {
					return child, nil //errors.New("Static already exists")
				}
				return child.Glue(node.Children[0])
			case len(child.Path) > max && len(node.Path) > max:
				_, err = child.Cut(max)
				if err != nil {
					return nil, err
				}
				node, err = child.Cut(max)
				if err != nil {
					return nil, err
				}
				child.Children = append(child.Children, node)
				node.SetRoot(child.Root())
				node.Parent = child
				return nil, nil
			case len(child.Path) > max:
				_, err = child.Cut(max)
				if err != nil {
					return nil, err
				}
				if !node.WildChild {
					return child, nil //errors.New("Static already exists")
				}
				node = node.Children[0]
				child.Children = append(child.Children, node)

				node.SetRoot(child.Root())
				node.Parent = child
				return nil, nil
			case len(node.Path) > max:
				node, err := node.Cut(max)
				if err != nil {
					return nil, err
				}
				return child.Glue(node)
			}
		}
	case Param:
		if n.Type == Param {
			return nil, errors.New("expected Static but actual Param")
		}
		for _, child := range n.Childrens(Param) {
			if bytes.Equal(node.Path, child.Path) {
				if !node.WildChild {
					return child, errors.New("Param already exists")
				}
				return child.Glue(node.Children[0])
			}
		}
	case CatchAll:
		for _, child := range n.Childrens(CatchAll) {
			return child, nil //errors.New("CatchAll already exists")
		}
	}
	n.Children = append(n.Children, node)
	node.SetRoot(n.Root())
	node.Parent = n
	return nil, nil
}

// Cut обрезает ноду на две части
func (n *Node) Cut(
	max int,
) (
	child *Node,
	err error,
) {
	switch {
	case max < 1:
		return nil, fmt.Errorf("expected max > 0 but actual %d", max)
	case n.Type != Static:
		return nil, fmt.Errorf("expected node type Static but actual %d", n.Type)
	case len(n.Path) < max:
		return nil, errors.New("expected len n.Path > max but actual n.Path < max")
	case len(n.Path) == max:
		return n, nil
	}
	child = &Node{
		root:     n.root,
		Parent:   n,
		Path:     n.Path[max:],
		Type:     Static,
		Children: n.Children,
	}
	n.Path = n.Path[:max]
	n.Children = []*Node{child}
	return child, nil
}

// SetRoot устанавливает ссылку на корневую ноду по всей цепочке
func (n *Node) SetRoot(root *Node) {
	n.root = root
	for _, node := range n.Children {
		node.SetRoot(root)
	}
}

// Identic возвращает количество совпадающих байт от начала адреса
func (n *Node) Identic(path []byte) (i int) {
	for i = 0; i < len(n.Path) && i < len(path); i++ {
		if n.Path[i] != path[i] {
			return i
		}
	}
	return i
}

// ChildIdentic возвращает ноду с максимальным совпадением
func (n *Node) ChildIdentic(
	path []byte,
) (
	max int,
	node *Node,
) {
	for _, child := range n.Childrens(Static) {
		identic := child.Identic(path)
		if identic > max {
			node = child
			max = identic
		}
	}
	return max, node
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
		case !bytes.HasPrefix(path[:len(node.Path)], node.Path):
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

// Set устанавливает значение по адресу
func (n *Node) Set(path []byte, v interface{}) error {
	lastNode, err := n.Insert(path)
	if err != nil {
		return err
	}
	lastNode.Value = v
	return nil
}
