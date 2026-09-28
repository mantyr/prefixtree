package prefixtree

// matchParamsSize число параметров, которое результат хранит в себе без отдельной аллокации
const matchParamsSize = 4

// Value это найденное по адресу значение
type Value interface {
	// Value возвращает сохранённое значение
	Value() interface{}

	// Path возвращает адрес запроса
	Path() string

	// Params возвращает найденные параметры
	Params() Params
}

// match это реализация Value
type match struct {
	node   *treeNode
	path   string
	params Params
	buf    [matchParamsSize]Param
}

// newMatch возвращает результат поиска с копией параметров
func newMatch(node *treeNode, path string, first, more Params) *match {
	m := &match{
		node: node,
		path: path,
	}
	size := len(first) + len(more)
	switch {
	case size == 0:
	case size <= matchParamsSize:
		n := copy(m.buf[:], first)
		m.params = m.buf[:n:n]
	default:
		m.params = make(Params, 0, size)
		m.params = append(m.params, first...)
		m.params = append(m.params, more...)
	}
	return m
}

// Value возвращает сохранённое значение
func (m *match) Value() interface{} {
	return m.node.value
}

// Path возвращает адрес запроса
func (m *match) Path() string {
	return m.path
}

// Params возвращает найденные параметры
func (m *match) Params() Params {
	return m.params
}
