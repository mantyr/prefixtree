package prefix

import (
	"bytes"
	"errors"
)

// First возвращает первую ноду в адресе
func First(
	path []byte,
) (
	node *Node,
	err error,
) {
	if len(path) == 0 {
		return nil, errors.New("empty path")
	}
	i := bytes.IndexAny(path, ":*")
	switch {
	case i < 0:
		// весь адрес - нода
		return &Node{
			Path: path,
		}, nil
	case i == 0:
		nType := Param
		if path[0] == '*' {
			nType = CatchAll
		}
		// первая нода - параметр, вычисляем длину
		if len(path[1:]) == 0 {
			return nil, errors.New("empty param name")
		}
		end := bytes.IndexAny(path[1:], "/:*")
		switch {
		case end < 0:
			// весь адрес - нода
			return &Node{
				Path: path[1:],
				Type: nType,
			}, nil
		case end == 0:
			return nil, errors.New("empty param name")
		}
		if nType == CatchAll {
			return nil, errors.New("unexpected '/'")
		}
		switch path[end+1] {
		case ':', '*':
			return nil, errors.New("expected '/' or end path but actual ':' or '*'")
		}
		return &Node{
			Path: path[1 : end+1],
			Type: Param,
		}, nil
	}
	// в адресе есть параметр идущий второй нодой
	return &Node{
		Path: path[:i],
	}, nil
}
