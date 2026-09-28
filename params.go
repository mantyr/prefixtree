package prefixtree

// Param это найденный параметр адреса
type Param struct {
	// Key имя параметра из шаблона без : и *
	Key string

	// Value значение параметра, подстрока адреса запроса
	Value string
}

// Params это найденные параметры в порядке следования в адресе
type Params []Param

// ByName возвращает значение параметра по имени
func (ps Params) ByName(name string) (string, bool) {
	for _, p := range ps {
		if p.Key == name {
			return p.Value, true
		}
	}
	return "", false
}

// Map возвращает параметры в виде map, аллоцирует на каждый вызов
func (ps Params) Map() map[string]string {
	m := make(map[string]string, len(ps))
	for _, p := range ps {
		m[p.Key] = p.Value
	}
	return m
}
