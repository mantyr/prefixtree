package prefixtree

import (
	"bytes"
	"errors"
	"fmt"
)

// Node это элемент параметризованного префиксного дерева
type Node struct {
	Token

	// root это ссылка на первый элемент в адресе
	root *Node

	// Parent это ссылка на вышестоящий элемент в адресе
	Parent *Node

	// Children это вложенные ноды
	Children []*Node

	// WildChild это флаг указывающий на наличие потомков
	WildChild bool

	// Value это хранимое в элементе дерева значение
	Value interface{}
}

// New возвращает новое дерево
func New() *Node {
	return &Node{
		Token: Token{
			Type: Root,
		},
	}
}

// SetString устанавливает значение по адресу
func (n *Node) SetString(path string, v interface{}) error {
	return n.Set([]byte(path), v)
}

// Set устанавливает значение по адресу
func (n *Node) Set(path []byte, v interface{}) error {
	if v == nil {
		return errors.New("empty value")
	}
	d := NewDecoder(path)
	tokens, err := d.Tokens()
	if err != nil {
		return err
	}
	node, err := n.Insert(tokens)
	if err != nil {
		return err
	}
	if node.Value != nil {
		return errors.New("path already in use")
	}
	node.Value = v
	return nil
}

// Insert вставляет в ветку токенов в дерево
func (n *Node) Insert(tokens Tokens) (*Node, error) {
	node := n
	var err error
	for _, token := range tokens {
		node, err = node.insert(token)
		if err != nil {
			return nil, err
		}
	}
	if n != node {
		return node, nil
	}
	return nil, errors.New("unexpected error")
}

// Insert сравнивает токен с текущими потомками
// возвращает последнюю ноду в образовавшейся цепочке
func (n *Node) insert(token Token) (*Node, error) {
	if n.Type == CatchAll {
		return nil, errors.New("expected EOF")
	}
	switch token.Type {
	case CatchAll:
		// если уже есть CatchAll то возвращаем ошибку, иначе просто добавляем ноду
		return n.insertCatchAll(token)
	case Param:
		// ищем полное совпадение
		return n.insertParam(token)
	case Static:
		// можно делить
		return n.insertStatic(token)
	}
	return nil, fmt.Errorf("unexpected token type %d", token.Type)
}

// insertCatchAll вставляет CatchAll токен
func (n *Node) insertCatchAll(token Token) (*Node, error) {
	switch n.Type {
	case Static, Root:
	default:
		return nil, errors.New("expected EOF")
	}
	for _, child := range n.Children {
		if child.Type == CatchAll {
			return nil, errors.New("CatchAll already exists")
		}
	}
	node := &Node{
		Token:  token,
		root:   n.root,
		Parent: n,
	}
	n.Children = append(n.Children, node)
	n.WildChild = true
	return node, nil
}

// insertParam вставляет Param токен
func (n *Node) insertParam(token Token) (*Node, error) {
	switch n.Type {
	case Static, Root:
	default:
		return nil, errors.New("expected EOF")
	}
	for _, child := range n.Children {
		if child.Type == Param && bytes.Equal(child.Title, token.Title) {
			return child, nil
		}
	}
	node := &Node{
		Token:  token,
		root:   n.root,
		Parent: n,
	}
	n.Children = append(n.Children, node)
	n.WildChild = true
	return node, nil
}

// insertStatic вставляет Static токен
func (n *Node) insertStatic(token Token) (*Node, error) {
	var node *Node
	var common []byte
	for _, child := range n.Children {
		if child.Type != Static {
			continue
		}
		prefix := CommonPrefix(child.Title, token.Title)
		if len(common) < len(prefix) {
			common = prefix
			node = child
		}
	}
	if node == nil {
		node = &Node{
			Token:  token,
			root:   n.root,
			Parent: n,
		}
		n.Children = append(n.Children, node)
		n.WildChild = true
		return node, nil
	}
	if len(common) < len(node.Title) {
		next := &Node{
			Token: Token{
				Title: node.Title[len(common):],
				Type:  Static,
			},
			root:      node.root,
			Parent:    node,
			Children:  node.Children,
			WildChild: node.WildChild,
			Value:     node.Value,
		}
		node.Title = node.Title[:len(common)]
		node.Children = []*Node{next}
		node.WildChild = true
		node.Value = nil
	}
	if bytes.Equal(node.Title, token.Title) {
		return node, nil
	}
	token.Title = token.Title[len(common):]
	return node.insertStatic(token)
}

// CommonPrefix возвращает общий префикс
func CommonPrefix(a, b []byte) []byte {
	var i int
	for i = 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			break
		}
	}
	return a[:i]
}

// GetString возвращает ноду со значением по определённому адресу
func (n *Node) GetString(path string) (*Value, error) {
	return n.Get([]byte(path))
}

// Get возвращает ноду со значеним по определённому адресу
func (n *Node) Get(path []byte) (*Value, error) {
	q := NewQuery(path)
	return n.get(q)
}

// Get возвращает ноду со значеним по определённому адресу
func (n *Node) get(q *Query) (*Value, error) {
	switch {
	case n.Type == CatchAll:
		return nil, errors.New("unexpected error - CatchAll")
	case !n.WildChild:
		return nil, errors.New("not found")
	case q.offset >= len(q.Path):
		return nil, errors.New("expected EOF")
	case n.Type == Param:
		return n.getStatic(q)
	}
	// Root, Static:
	value, err := n.getStatic(q)
	if err == nil {
		return value, nil
	}
	value, err = n.getParam(q)
	if err == nil {
		return value, nil
	}
	return n.getCatchAll(q)
}

// getStatic возвращет цепочку где первая нода Static
func (n *Node) getStatic(q *Query) (*Value, error) {
	path := q.path()
	for _, child := range n.Children {
		if child.Type != Static {
			continue
		}
		if len(child.Title) > len(path) {
			continue
		}
		if bytes.Equal(child.Title, path[:len(child.Title)]) {
			if len(child.Title) == len(path) {
				if child.Value == nil {
					return nil, errors.New("not found")
				}
				return NewValue(q, child), nil
			}
			q.offset += len(child.Title)
			value, err := child.get(q)
			if err != nil {
				q.offset -= len(child.Title)
			}
			return value, err
		}
	}
	return nil, errors.New("not found")
}

// getParam возвращает цепочку где первая нода Param
func (n *Node) getParam(q *Query) (*Value, error) {
	path := q.path()
	end := bytes.IndexAny(path, "/")
	switch {
	case end < 0:
		for _, child := range n.Children {
			if child.Type == Param && child.Value != nil {
				q.Params[string(child.Title)] = string(path)
				return NewValue(q, child), nil
			}
		}
		return nil, errors.New("not found")
	}
	// после значения параметра есть Static токен
	for _, child := range n.Children {
		if child.Type != Param {
			continue
		}
		q.offset += end
		q.Params[string(child.Title)] = string(path[:end])
		value, err := child.get(q)
		if err == nil {
			return value, nil
		}
		delete(q.Params, string(child.Title))
		q.offset -= end
	}
	return nil, errors.New("not found")
}

// getCatchAll возвращает цепочку где первая нода CatchAll
func (n *Node) getCatchAll(q *Query) (*Value, error) {
	for _, child := range n.Children {
		if child.Type != CatchAll {
			continue
		}
		if child.Value == nil {
			continue
		}
		q.Params[string(child.Title)] = string(q.path())
		return NewValue(q, child), nil
	}
	return nil, errors.New("not found")
}
