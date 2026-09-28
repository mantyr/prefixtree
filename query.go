package prefixtree

// queryParamsSize число параметров, которое поиск держит без аллокаций
const queryParamsSize = 8

// query это состояние поиска по дереву, живёт только внутри Get
type query struct {
	// offset позиция в path, до которой адрес уже разобран
	offset int

	// path адрес запроса
	path string

	// size число найденных параметров, при откате возвращается к прежнему
	size int

	// buf первые параметры, хранятся без аллокаций
	buf [queryParamsSize]Param

	// more параметры сверх buf
	more Params
}

// rest возвращает необработанный остаток адреса
func (q *query) rest() string {
	return q.path[q.offset:]
}

// push добавляет найденный параметр
func (q *query) push(key, value string) {
	p := Param{Key: key, Value: value}
	if q.size < queryParamsSize {
		q.buf[q.size] = p
	} else {
		q.more = append(q.more[:q.size-queryParamsSize], p)
	}
	q.size++
}

// truncate откатывает параметры до указанного числа
func (q *query) truncate(size int) {
	q.size = size
}

// params возвращает найденные параметры, first и more не пересекаются
func (q *query) params() (first, more Params) {
	if q.size <= queryParamsSize {
		return q.buf[:q.size], nil
	}
	return q.buf[:], q.more[:q.size-queryParamsSize]
}
