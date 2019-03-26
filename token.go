package prefixtree

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

// Перечень видов токена
const (
	Root = iota
	Static
	CatchAll
	Param
)

// TokenType это тип токена
type TokenType uint8

// Token это самостоятельная единица
type Token struct {
	// Type это тип токена
	Type TokenType

	// Title это название токена
	Title []byte
}

// Tokens это набор токенов
type Tokens []Token

// String реализует интерфейс fmt.Stringer
func (t *Tokens) String() string {
	var title string
	for _, token := range []Token(*t) {
		title = title + token.String()
	}
	return title
}

// String реализует интерфейс fmt.Stringer
func (t *Token) String() string {
	switch t.Type {
	case Root:
		return "^"
	case Static:
		return "[" + string(t.Title) + "]"
	case CatchAll:
		return "*" + string(t.Title)
	case Param:
		return ":" + string(t.Title)
	}
	return "?" + string(t.Title)
}

// Decoder разбирает набор байт на токены
type Decoder struct {
	offset int
	max    int
	data   []byte
	state  TokenType
}

// NewDecoder возвращает новый декодер адресов
func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		max:  len(data),
		data: data,
	}
}

// Tokens возвращает набор токенов и ошибку в случае если не удалось распарсить
func (d *Decoder) Tokens() (Tokens, error) {
	var tokens Tokens
	var token *Token
	err := d.PathValid()
	if err != nil {
		return tokens, err
	}
	for {
		token, err = d.Token()
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

// PathValid проверяет отсутствие запрещённых символов в path
func (d *Decoder) PathValid() error {
	unvalid := bytes.IndexAny(d.data, "\r\t\n?= ")
	if unvalid < 0 {
		return nil
	}
	return fmt.Errorf(`unexpected char "%c"`, d.data[unvalid])
}

// Token возвращает следующией токен
func (d *Decoder) Token() (token *Token, err error) {
	if d.offset >= d.max {
		return nil, io.EOF
	}
	if d.state == CatchAll {
		return nil, errors.New("expected EOF")
	}
	token = &Token{}
	switch d.data[d.offset] {
	case ':', '*':
		switch d.state {
		case Static, Root:
		default:
			return nil, errors.New("expected Static token")
		}
	}
	switch d.data[d.offset] {
	case ':':
		token.Type = Param
		d.state = Param
		d.offset++
	case '*':
		token.Type = CatchAll
		d.state = CatchAll
		d.offset++
	default:
		token.Type = Static
		d.state = Static
	}
	token.Title, err = d.parse()
	if err != nil {
		return nil, err
	}
	return token, nil
}

// parse возвращает название токена
func (d *Decoder) parse() ([]byte, error) {
	if d.offset >= d.max {
		return nil, errors.New("empty token value")
	}
	var end int

	switch d.state {
	case CatchAll, Param:
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
		return nil, errors.New("empty token value")
	}
}
