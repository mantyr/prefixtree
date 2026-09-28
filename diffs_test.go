package prefixtree

import (
	"fmt"
)

// diffsLimit сколько расхождений показывать в сообщении о падении
const diffsLimit = 10

// diffs копит расхождения в переборных тестах, чтобы проверять их одним So в конце
type diffs struct {
	items []string
	total int
}

// equal запоминает расхождение, если got и want не совпали
func (d *diffs) equal(label string, got, want interface{}) {
	g, w := fmt.Sprint(got), fmt.Sprint(want)
	if g == w {
		return
	}
	d.add(fmt.Sprintf("%s\n    got:  %s\n    want: %s", label, g, w))
}

// add запоминает расхождение
func (d *diffs) add(item string) {
	d.total++
	if len(d.items) < diffsLimit {
		d.items = append(d.items, item)
	}
}

// addAll запоминает список расхождений
func (d *diffs) addAll(label string, items []string) {
	for _, item := range items {
		d.add(label + ": " + item)
	}
}

// result возвращает первые расхождения и их общее число, пустая строка если расхождений нет
func (d *diffs) result() string {
	if d.total == 0 {
		return ""
	}
	s := fmt.Sprintf("%d mismatches, first %d:", d.total, len(d.items))
	for _, item := range d.items {
		s += "\n  " + item
	}
	return s
}
