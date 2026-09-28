package prefixtree

import (
	"bytes"
	"io"
)

// Перечень видов токена
const (
	rootToken tokenKind = iota
	staticToken
	catchAllToken
	paramToken
)

// tokenKind это тип токена
type tokenKind uint8

// pathToken это самостоятельная единица
type pathToken struct {
	// kind это тип токена
	kind tokenKind

	// title это название токена
	title []byte
}

// pathTokens это набор токенов
type pathTokens []pathToken

// checkParamNames проверяет что имена paramToken и catchAllToken не повторяются в одном шаблоне
func (t pathTokens) checkParamNames() error {
	for i, token := range t {
		if token.kind != paramToken && token.kind != catchAllToken {
			continue
		}
		for _, next := range t[i+1:] {
			if next.kind != paramToken && next.kind != catchAllToken {
				continue
			}
			if bytes.Equal(token.title, next.title) {
				return newError(DuplicateParamInPath, `duplicate param "%s"`, token.title)
			}
		}
	}
	return nil
}

// view возвращает текстовое представление набора токенов
func (t *pathTokens) view() string {
	var title string
	for _, token := range []pathToken(*t) {
		title = title + token.view()
	}
	return title
}

// view возвращает текстовое представление токена
func (t *pathToken) view() string {
	switch t.kind {
	case rootToken:
		return "^"
	case staticToken:
		return "[" + string(t.title) + "]"
	case catchAllToken:
		return "*" + string(t.title)
	case paramToken:
		return ":" + string(t.title)
	}
	return "?" + string(t.title)
}

// decoder разбирает набор байт на токены
type decoder struct {
	offset int
	max    int
	data   []byte
	state  tokenKind
}

// newDecoder возвращает новый декодер адресов
func newDecoder(data []byte) *decoder {
	return &decoder{
		max:  len(data),
		data: data,
	}
}

// tokens возвращает набор токенов и ошибку в случае если не удалось распарсить
func (d *decoder) tokens() (pathTokens, error) {
	var tokens pathTokens
	var token *pathToken
	err := d.pathValid()
	if err != nil {
		return tokens, err
	}
	for {
		token, err = d.nextToken()
		if err == nil {
			tokens = append(tokens, *token)
			continue
		}
		if err == io.EOF {
			return tokens, nil
		}
		return tokens, err
	}
}

// pathValid проверяет отсутствие запрещённых символов в path
func (d *decoder) pathValid() error {
	unvalid := bytes.IndexAny(d.data, "\r\t\n?= ")
	if unvalid < 0 {
		return nil
	}
	return newError(InvalidPath, `unexpected char "%c"`, d.data[unvalid])
}

// nextToken возвращает следующий токен
func (d *decoder) nextToken() (token *pathToken, err error) {
	if d.offset >= d.max {
		return nil, io.EOF
	}
	if d.state == catchAllToken {
		return nil, errExpectedEOF
	}
	token = &pathToken{}
	switch d.data[d.offset] {
	case ':', '*':
		switch d.state {
		case staticToken, rootToken:
		default:
			return nil, errExpectedStaticToken
		}
	}
	switch d.data[d.offset] {
	case ':':
		token.kind = paramToken
		d.state = paramToken
		d.offset++
	case '*':
		token.kind = catchAllToken
		d.state = catchAllToken
		d.offset++
	default:
		token.kind = staticToken
		d.state = staticToken
	}
	token.title, err = d.parse()
	if err != nil {
		return nil, err
	}
	return token, nil
}

// parse возвращает название токена
func (d *decoder) parse() ([]byte, error) {
	if d.offset >= d.max {
		return nil, errEmptyTokenValue
	}
	var end int

	switch d.state {
	case catchAllToken, paramToken:
		end = bytes.IndexAny(d.data[d.offset:], "/:*")
	default:
		end = bytes.IndexAny(d.data[d.offset:], ":*")
	}

	switch {
	case end < 0:
		data := d.data[d.offset:]
		d.offset = d.max
		return data, nil
	case end > 0:
		data := d.data[d.offset : d.offset+end]
		d.offset = d.offset + end
		return data, nil
	default:
		return nil, errEmptyTokenValue
	}
}
