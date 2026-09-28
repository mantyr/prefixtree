package prefixtree

import (
	"testing"
)

// benchTree возвращает дерево с маршрутами для бенчмарков
func benchTree(b *testing.B) *treeNode {
	root := newNode()
	for _, path := range []string{
		"/s/*path",
		"/u/:slot/oauth/*path",
		"/u/:slot/api/*path",
		"/api/v1/*path",
		"/api/v1/health",
	} {
		if err := root.SetString(path, path); err != nil {
			b.Fatal(err)
		}
	}
	return root
}

func BenchmarkGetParams(b *testing.B) {
	root := benchTree(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root.GetString("/u/abc/api/users/1") //nolint:errcheck
	}
}

func BenchmarkGetStatic(b *testing.B) {
	root := benchTree(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root.GetString("/api/v1/health") //nolint:errcheck
	}
}

func BenchmarkGetNotFound(b *testing.B) {
	root := benchTree(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root.GetString("/u/abc/other/x") //nolint:errcheck
	}
}

func BenchmarkClone(b *testing.B) {
	root := newNode()
	for _, path := range deleteRoutes {
		if err := root.SetString(path, path); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root.clone(nil)
	}
}

func BenchmarkRebuild(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		root := newNode()
		for _, path := range deleteRoutes {
			root.SetString(path, path) //nolint:errcheck
		}
	}
}

func BenchmarkStorageGet(b *testing.B) {
	s := NewStorage()
	for _, path := range deleteRoutes {
		if err := s.Set(path, path); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Tree().GetString("/u/abc/api/users/1") //nolint:errcheck
	}
}
