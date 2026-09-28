package prefixtree

import (
	"bytes"
	"reflect"
	"strings"
)

// treeNode это элемент параметризованного префиксного дерева, наружу отдаётся как Tree
type treeNode struct {
	pathToken

	// parent это ссылка на вышестоящий элемент в адресе
	parent *treeNode

	// children это вложенные ноды
	children []*treeNode

	// wildChild это флаг указывающий на наличие потомков
	wildChild bool

	// value это хранимое в элементе дерева значение
	value interface{}

	// key имя параметра для paramToken и catchAllToken, заводится один раз при вставке
	key string
}

// New возвращает новое пустое дерево
func New() Tree {
	return newNode()
}

// newNode возвращает корень нового дерева
func newNode() *treeNode {
	return &treeNode{
		pathToken: pathToken{
			kind: rootToken,
		},
	}
}

// SetString устанавливает значение по адресу
func (n *treeNode) SetString(path string, v interface{}) error {
	return n.set([]byte(path), v)
}

// Set устанавливает значение по адресу, path копируется
func (n *treeNode) Set(path []byte, v interface{}) error {
	return n.set(bytes.Clone(path), v)
}

// set устанавливает значение, path должен принадлежать дереву
func (n *treeNode) set(path []byte, v interface{}) error {
	if v == nil {
		return errEmptyValue
	}
	d := newDecoder(path)
	tokens, err := d.tokens()
	if err != nil {
		return err
	}
	node, err := n.insertTokens(tokens)
	if err != nil {
		return err
	}
	if node.value != nil {
		return errPathInUse
	}
	node.value = v
	return nil
}

// DeleteString удаляет значение по шаблону адреса
func (n *treeNode) DeleteString(path string) error {
	return n.delete([]byte(path))
}

// Delete удаляет значение по шаблону адреса
func (n *treeNode) Delete(path []byte) error {
	return n.delete(path)
}

// delete удаляет значение и вычищает ставшие пустыми ноды
func (n *treeNode) delete(path []byte) error {
	tokens, err := patternTokens(path)
	if err != nil {
		return err
	}
	node := n.lookup(tokens)
	if node == nil || node.value == nil {
		return errNotFound
	}
	node.value = nil
	node.prune()
	return nil
}

// ReplaceString заменяет шаблон oldPath на newPath со значением v, при совпадении шаблонов меняет только значение
// либо возвращает ошибку и не меняет дерево, либо выполняется целиком
// тот же шаблон с тем же значением (тот же указатель) даёт NotChanged
func (n *treeNode) ReplaceString(oldPath, newPath string, v interface{}) error {
	return n.replace([]byte(oldPath), []byte(newPath), v)
}

// Replace заменяет шаблон oldPath на newPath со значением v, newPath копируется
func (n *treeNode) Replace(oldPath, newPath []byte, v interface{}) error {
	return n.replace(oldPath, bytes.Clone(newPath), v)
}

// replace заменяет шаблон, newPath должен принадлежать дереву
func (n *treeNode) replace(oldPath, newPath []byte, v interface{}) error {
	if v == nil {
		return errEmptyValue
	}
	oldTokens, err := patternTokens(oldPath)
	if err != nil {
		return err
	}
	oldNode := n.lookup(oldTokens)
	if oldNode == nil || oldNode.value == nil {
		return errNotFound
	}
	if bytes.Equal(oldPath, newPath) {
		if sameValue(oldNode.value, v) {
			return errNotChanged
		}
		oldNode.value = v
		return nil
	}
	newTokens, err := patternTokens(newPath)
	if err != nil {
		return err
	}
	err = newTokens.checkParamNames()
	if err != nil {
		return err
	}
	// все проверки до изменений, чтобы ошибка не оставила дерево без old
	err = n.canInsert(newTokens, oldNode)
	if err != nil {
		return err
	}
	oldNode.value = nil
	oldNode.prune()
	node, err := n.insertTokens(newTokens)
	mustInsert(err)
	node.value = v
	return nil
}

// sameValue сравнивает значения без паники, несравнимые (func, map, slice и содержащие их) считаются разными
func sameValue(a, b interface{}) bool {
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
	if va.Type() != vb.Type() || !va.Comparable() || !vb.Comparable() {
		return false
	}
	return a == b
}

// patternTokens разбирает шаблон, пустой шаблон это ошибка
func patternTokens(path []byte) (pathTokens, error) {
	tokens, err := newDecoder(path).tokens()
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, errEmptyPath
	}
	return tokens, nil
}

// canInsert проверяет без изменений, что шаблон можно вставить, ignore это нода, которая будет удалена до вставки
func (n *treeNode) canInsert(tokens pathTokens, ignore *treeNode) error {
	node := n
	for _, token := range tokens {
		switch token.kind {
		case staticToken:
			for title := token.title; len(title) > 0; {
				next := node.staticPrefixChild(title)
				if next == nil {
					// дальше пойдут новые ноды, конфликтовать не с чем
					return nil
				}
				title = title[len(next.title):]
				node = next
			}
		case paramToken:
			next := node.lookupChild(token)
			if next == nil {
				return nil
			}
			node = next
		case catchAllToken:
			var next *treeNode
			for _, child := range node.children {
				if child.kind != catchAllToken || child == ignore {
					continue
				}
				if !bytes.Equal(child.title, token.title) {
					return errCatchAllExists
				}
				next = child
			}
			if next == nil {
				return nil
			}
			node = next
		}
	}
	if node != ignore && node.value != nil {
		return errPathInUse
	}
	return nil
}

// staticPrefixChild возвращает статического потомка, заголовок которого целиком входит в начало title
func (n *treeNode) staticPrefixChild(title []byte) *treeNode {
	for _, child := range n.children {
		if child.kind != staticToken || len(child.title) > len(title) {
			continue
		}
		if bytes.Equal(child.title, title[:len(child.title)]) {
			return child
		}
	}
	return nil
}

// mustInsert падает, если вставка после canInsert всё же не удалась: это нарушенный инвариант дерева
func mustInsert(err error) {
	if err != nil {
		panic(err)
	}
}

// lookup возвращает ноду точно по токенам шаблона или nil
func (n *treeNode) lookup(tokens pathTokens) *treeNode {
	node := n
	for _, token := range tokens {
		if token.kind == staticToken {
			node = node.lookupStatic(token.title)
		} else {
			node = node.lookupChild(token)
		}
		if node == nil {
			return nil
		}
	}
	return node
}

// lookupStatic спускается по статическим нодам, которые вместе дают title
func (n *treeNode) lookupStatic(title []byte) *treeNode {
	node := n
	for len(title) > 0 {
		next := node.staticPrefixChild(title)
		if next == nil {
			return nil
		}
		title = title[len(next.title):]
		node = next
	}
	return node
}

// lookupChild возвращает потомка с тем же типом и именем, что у токена
func (n *treeNode) lookupChild(token pathToken) *treeNode {
	for _, child := range n.children {
		if child.kind == token.kind && bytes.Equal(child.title, token.title) {
			return child
		}
	}
	return nil
}

// prune удаляет пустые ноды вверх по дереву и склеивает оставшуюся статику
func (n *treeNode) prune() {
	node := n
	for node.parent != nil && node.value == nil && len(node.children) == 0 {
		parent := node.parent
		parent.removeChild(node)
		node = parent
	}
	node.merge()
}

// removeChild отцепляет потомка
func (n *treeNode) removeChild(child *treeNode) {
	for i, c := range n.children {
		if c == child {
			n.children = append(n.children[:i], n.children[i+1:]...)
			break
		}
	}
	n.wildChild = len(n.children) > 0
	child.parent = nil
}

// merge склеивает статическую ноду без значения с единственным статическим потомком
func (n *treeNode) merge() {
	for n.kind == staticToken && n.value == nil && len(n.children) == 1 {
		child := n.children[0]
		if child.kind != staticToken {
			return
		}
		// полное выражение среза, чтобы не писать в чужой массив после деления ноды
		n.title = append(n.title[:len(n.title):len(n.title)], child.title...)
		n.children = child.children
		n.wildChild = child.wildChild
		n.value = child.value
		for _, c := range n.children {
			c.parent = n
		}
	}
}

// insertTokens вставляет ветку токенов в дерево
func (n *treeNode) insertTokens(tokens pathTokens) (*treeNode, error) {
	err := tokens.checkParamNames()
	if err != nil {
		return nil, err
	}
	node := n
	for _, token := range tokens {
		node, err = node.insert(token)
		if err != nil {
			return nil, err
		}
	}
	if n != node {
		return node, nil
	}
	return nil, errEmptyPath
}

// insert сравнивает токен с текущими потомками
// возвращает последнюю ноду в образовавшейся цепочке
func (n *treeNode) insert(token pathToken) (*treeNode, error) {
	if n.kind == catchAllToken {
		return nil, errExpectedEOF
	}
	switch token.kind {
	case catchAllToken:
		// если уже есть catchAllToken то возвращаем ошибку, иначе просто добавляем ноду
		return n.insertCatchAll(token)
	case paramToken:
		// ищем полное совпадение
		return n.insertParam(token)
	case staticToken:
		// можно делить
		return n.insertStatic(token)
	}
	return nil, newError(Internal, "unexpected token type %d", token.kind)
}

// insertCatchAll вставляет catchAllToken токен
func (n *treeNode) insertCatchAll(token pathToken) (*treeNode, error) {
	switch n.kind {
	case staticToken, rootToken:
	default:
		return nil, errExpectedEOF
	}
	for _, child := range n.children {
		if child.kind == catchAllToken {
			return nil, errCatchAllExists
		}
	}
	node := &treeNode{
		pathToken: token,
		key:       string(token.title),
		parent:    n,
	}
	n.children = append(n.children, node)
	n.wildChild = true
	return node, nil
}

// insertParam вставляет paramToken токен
func (n *treeNode) insertParam(token pathToken) (*treeNode, error) {
	switch n.kind {
	case staticToken, rootToken:
	default:
		return nil, errExpectedEOF
	}
	for _, child := range n.children {
		if child.kind == paramToken && bytes.Equal(child.title, token.title) {
			return child, nil
		}
	}
	node := &treeNode{
		pathToken: token,
		key:       string(token.title),
		parent:    n,
	}
	n.children = append(n.children, node)
	n.wildChild = true
	return node, nil
}

// insertStatic вставляет staticToken токен
func (n *treeNode) insertStatic(token pathToken) (*treeNode, error) {
	var node *treeNode
	var common []byte
	for _, child := range n.children {
		if child.kind != staticToken {
			continue
		}
		prefix := commonPrefix(child.title, token.title)
		if len(common) < len(prefix) {
			common = prefix
			node = child
		}
	}
	if node == nil {
		node = &treeNode{
			pathToken: token,
			parent:    n,
		}
		n.children = append(n.children, node)
		n.wildChild = true
		return node, nil
	}
	if len(common) < len(node.title) {
		next := &treeNode{
			pathToken: pathToken{
				title: node.title[len(common):],
				kind:  staticToken,
			},
			parent:    node,
			children:  node.children,
			wildChild: node.wildChild,
			value:     node.value,
		}
		// перенесённые потомки теперь висят на next
		for _, child := range next.children {
			child.parent = next
		}
		node.title = node.title[:len(common)]
		node.children = []*treeNode{next}
		node.wildChild = true
		node.value = nil
	}
	if bytes.Equal(node.title, token.title) {
		return node, nil
	}
	token.title = token.title[len(common):]
	return node.insertStatic(token)
}

// commonPrefix возвращает общий префикс
func commonPrefix(a, b []byte) []byte {
	var i int
	for i = 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			break
		}
	}
	return a[:i]
}

// GetString возвращает значение по адресу
func (n *treeNode) GetString(path string) (Value, error) {
	q := query{path: path}
	node := n.find(&q)
	if node == nil {
		return nil, errNotFound
	}
	first, more := q.params()
	return newMatch(node, path, first, more), nil
}

// Get возвращает значение по адресу
func (n *treeNode) Get(path []byte) (Value, error) {
	return n.GetString(string(path))
}

// find возвращает ноду со значением или nil
func (n *treeNode) find(q *query) *treeNode {
	switch {
	case n.kind == catchAllToken:
		return nil
	case !n.wildChild:
		return nil
	case q.offset >= len(q.path):
		return nil
	case n.kind == paramToken:
		return n.findStatic(q)
	}
	// rootToken, staticToken:
	if node := n.findStatic(q); node != nil {
		return node
	}
	if node := n.findParam(q); node != nil {
		return node
	}
	return n.findCatchAll(q)
}

// findStatic ищет цепочку где первая нода staticToken
func (n *treeNode) findStatic(q *query) *treeNode {
	path := q.rest()
	for _, child := range n.children {
		if child.kind != staticToken {
			continue
		}
		if len(child.title) > len(path) {
			continue
		}
		if string(child.title) != path[:len(child.title)] {
			continue
		}
		if len(child.title) == len(path) {
			if child.value == nil {
				return nil
			}
			return child
		}
		q.offset += len(child.title)
		node := child.find(q)
		if node == nil {
			q.offset -= len(child.title)
		}
		return node
	}
	return nil
}

// findParam ищет цепочку где первая нода paramToken
func (n *treeNode) findParam(q *query) *treeNode {
	path := q.rest()
	end := strings.IndexByte(path, '/')
	if end < 0 {
		for _, child := range n.children {
			if child.kind == paramToken && child.value != nil {
				q.push(child.key, path)
				return child
			}
		}
		return nil
	}
	// после значения параметра есть staticToken токен
	size := q.size
	for _, child := range n.children {
		if child.kind != paramToken {
			continue
		}
		q.offset += end
		q.push(child.key, path[:end])
		if node := child.find(q); node != nil {
			return node
		}
		q.truncate(size)
		q.offset -= end
	}
	return nil
}

// findCatchAll ищет цепочку где первая нода catchAllToken
func (n *treeNode) findCatchAll(q *query) *treeNode {
	for _, child := range n.children {
		if child.kind != catchAllToken || child.value == nil {
			continue
		}
		q.push(child.key, q.rest())
		return child
	}
	return nil
}
