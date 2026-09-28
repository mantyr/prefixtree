package prefixtree

import (
	"github.com/mantyr/codes"
)

// Стандартные коды, которые может вернуть пакет
const (
	// OK — ошибки нет
	OK = codes.OK

	// Unknown — ошибка не из этого пакета
	Unknown = codes.Unknown

	// Internal — нарушен инвариант дерева, баг пакета
	Internal = codes.Internal
)

// Коды пакета
const (
	// PathNotFound — по адресу нет значения
	PathNotFound codes.Code = iota + codes.CustomCodes + 1

	// PathAlreadyExists — шаблон уже занят другим значением
	PathAlreadyExists

	// InvalidPath — шаблон адреса не удалось разобрать
	InvalidPath

	// EmptyValue — попытка сохранить nil
	EmptyValue

	// DuplicateParamInPath — имя параметра повторяется в одном шаблоне
	DuplicateParamInPath

	// NotChanged — Replace на тот же шаблон с тем же значением, дерево не менялось
	NotChanged
)

// Code возвращает код ошибки
func Code(err error) codes.Code {
	return codes.CodeOf(err)
}

// newError возвращает ошибку с кодом пакета
func newError(code codes.Code, args ...interface{}) error {
	return codes.NewError(code, args...)
}
