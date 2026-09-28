package prefixtree

import (
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetConcurrent(t *testing.T) {
	Convey("Параллельные Get на готовом дереве", t, func() {
		root := newNode()
		So(root.SetString("/u/:slot/api/*path", "uapi"), ShouldBeNil)
		So(root.SetString("/u/:slot/oauth/*path", "oauth"), ShouldBeNil)

		var wg sync.WaitGroup
		errs := make(chan string, 64)
		for i := 0; i < 64; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 200; j++ {
					value, err := root.GetString("/u/abc/oauth/x")
					if err != nil || value.Value() != "oauth" {
						errs <- "bad result"
						return
					}
					if slot, _ := value.Params().ByName("slot"); slot != "abc" {
						errs <- "bad slot"
						return
					}
				}
			}()
		}
		wg.Wait()
		close(errs)
		So(len(errs), ShouldEqual, 0)
	})
}
