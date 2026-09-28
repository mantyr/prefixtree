package prefixtree

import (
	"sync"
	"sync/atomic"
)

// TreeReadonly это дерево только для чтения
type TreeReadonly interface {
	// Get возвращает значение по адресу
	Get(path []byte) (Value, error)

	// GetString возвращает значение по адресу
	GetString(path string) (Value, error)
}

// Tree это префиксное дерево со всеми операциями, изменения нельзя делать параллельно с другими вызовами
type Tree interface {
	TreeReadonly

	// Set устанавливает значение по шаблону адреса, path копируется
	Set(path []byte, v interface{}) error

	// SetString устанавливает значение по шаблону адреса
	SetString(path string, v interface{}) error

	// Delete удаляет значение по шаблону адреса
	Delete(path []byte) error

	// DeleteString удаляет значение по шаблону адреса
	DeleteString(path string) error

	// Replace заменяет шаблон oldPath на newPath со значением v, newPath копируется
	Replace(oldPath, newPath []byte, v interface{}) error

	// ReplaceString заменяет шаблон oldPath на newPath со значением v, без изменений возвращает NotChanged
	ReplaceString(oldPath, newPath string, v interface{}) error

	// Clone возвращает полную копию дерева
	Clone() Tree
}

// проверка, что treeNode реализует интерфейсы
var (
	_ Tree         = (*treeNode)(nil)
	_ TreeReadonly = (*treeNode)(nil)
)

// Storage это дерево для конкурентного чтения: изменения идут в мастер под мьютексом, читатели берут копию из атомика
type Storage interface {
	// Set устанавливает значение и публикует копию дерева
	Set(path string, v interface{}) error

	// Delete удаляет значение и публикует копию дерева
	Delete(path string) error

	// Replace заменяет шаблон и публикует копию дерева, при NotChanged ничего не публикует
	Replace(oldPath, newPath string, v interface{}) error

	// Tree возвращает опубликованную копию дерева только для чтения
	Tree() TreeReadonly
}

// storage это реализация Storage
type storage struct {
	// mu сериализует изменения, читатели его не берут
	mu sync.Mutex

	// master изменяемое дерево, трогается только под mu
	master *treeNode

	// current опубликованная копия master, после публикации не меняется
	current atomic.Pointer[treeNode]
}

// NewStorage возвращает пустое хранилище
func NewStorage() Storage {
	s := &storage{
		master: newNode(),
	}
	s.current.Store(newNode())
	return s
}

// Set устанавливает значение и публикует копию дерева
// при ошибке мастер не тронут и ничего не публикуется
func (s *storage) Set(path string, v interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.master.SetString(path, v)
	if err != nil {
		return err
	}
	s.current.Store(s.master.clone(nil))
	return nil
}

// Delete удаляет значение и публикует копию дерева
func (s *storage) Delete(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.master.DeleteString(path)
	if err != nil {
		return err
	}
	s.current.Store(s.master.clone(nil))
	return nil
}

// Replace заменяет шаблон и публикует копию дерева, при NotChanged ничего не публикует
func (s *storage) Replace(oldPath, newPath string, v interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.master.replace([]byte(oldPath), []byte(newPath), v)
	if err != nil {
		return err
	}
	s.current.Store(s.master.clone(nil))
	return nil
}

// Tree возвращает опубликованную копию дерева только для чтения
func (s *storage) Tree() TreeReadonly {
	return s.current.Load()
}

// Clone возвращает полную копию дерева, которую можно отдать читателям, пока оригинал меняется
// байты title разделяются с оригиналом: дерево никогда не пишет в уже существующие байты, merge дописывает в новый массив
func (n *treeNode) Clone() Tree {
	return n.clone(nil)
}

// clone возвращает копию поддерева с указанным родителем
func (n *treeNode) clone(parent *treeNode) *treeNode {
	c := &treeNode{
		pathToken: n.pathToken,
		parent:    parent,
		wildChild: n.wildChild,
		value:     n.value,
		key:       n.key,
	}
	if len(n.children) > 0 {
		c.children = make([]*treeNode, len(n.children))
		for i, child := range n.children {
			c.children[i] = child.clone(c)
		}
	}
	return c
}
