package prefixtree

// Query это запрос с хранением найденных параметров
type Query struct {
	offset int
	Path   []byte
	Params map[string]string
}

// Value это значение запроса с дополнительной информацией
type Value struct {
	*Query
	*Node
}

// NewQuery возвращает новый запрос
func NewQuery(path []byte) *Query {
	return &Query{
		Path:   path,
		Params: make(map[string]string),
	}
}

// path возвращает необработанный остаток адреса
func (q *Query) path() []byte {
	return q.Path[q.offset:]
}

// NewValue возвращает новое значение запроса
func NewValue(q *Query, n *Node) *Value {
	return &Value{
		Query: q,
		Node:  n,
	}
}
