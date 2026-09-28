package prefixtree_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/mantyr/prefixtree/v2"
)

// router это HTTP-роутер поверх Storage: маршруты можно менять на лету, запросы не ждут
type router struct {
	routes prefixtree.Storage
}

// handle это обработчик с параметрами адреса
type handle func(w http.ResponseWriter, r *http.Request, params prefixtree.Params)

// ServeHTTP находит обработчик по адресу запроса
func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	value, err := rt.routes.Tree().GetString(r.URL.Path)
	switch prefixtree.Code(err) {
	case prefixtree.OK:
		value.Value().(handle)(w, r, value.Params())
	case prefixtree.PathNotFound:
		http.NotFound(w, r)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Example_router() {
	rt := &router{routes: prefixtree.NewStorage()}

	var api handle = func(w http.ResponseWriter, r *http.Request, params prefixtree.Params) {
		slot, _ := params.ByName("slot")
		path, _ := params.ByName("path")
		fmt.Fprintf(w, "api slot=%s path=%s", slot, path) //nolint:errcheck
	}
	var apiV2 handle = func(w http.ResponseWriter, r *http.Request, params prefixtree.Params) {
		fmt.Fprint(w, "api v2") //nolint:errcheck
	}
	if err := rt.routes.Set("/u/:slot/api/*path", api); err != nil {
		fmt.Println(err)
	}

	get := func(path string) {
		w := httptest.NewRecorder()
		rt.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		fmt.Println(w.Code, strings.TrimSpace(w.Body.String()))
	}

	get("/u/abc/api/users/1")
	get("/nope")

	// переезд маршрута на лету, запросы видят либо старую, либо новую версию
	if err := rt.routes.Replace("/u/:slot/api/*path", "/u/:slot/v2/*path", apiV2); err != nil {
		fmt.Println(err)
	}
	get("/u/abc/api/users/1")
	get("/u/abc/v2/users/1")
	// Output:
	// 200 api slot=abc path=users/1
	// 404 404 page not found
	// 404 404 page not found
	// 200 api v2
}
