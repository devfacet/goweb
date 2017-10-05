/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package route_test

import (
	"net/http"
	"sort"
	"testing"

	"github.com/devfacet/goweb/route"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("should return a new route", t, func() {
		r := route.New(route.Options{Path: "/test"})
		So(r.Path(), ShouldEqual, "/test")
		So(r.Pattern(), ShouldEqual, "/test")
		So(r.Explicit(), ShouldEqual, true)
		So(r.Redirect(), ShouldEqual, false)
	})
}

func TestPath(t *testing.T) {
	Convey("should return the given path value", t, func() {
		r := route.New(route.Options{Path: "/test"})
		So(r.Path(), ShouldEqual, "/test")
	})
}

func TestPattern(t *testing.T) {
	Convey("should return the correct pattern value", t, func() {
		r := route.New(route.Options{Path: "/test"})
		So(r.Pattern(), ShouldEqual, "/test")
	})
}

func TestExplicit(t *testing.T) {
	Convey("should return the correct explicit value", t, func() {
		r := route.New(route.Options{Path: "/test"})
		So(r.Explicit(), ShouldEqual, true)
	})
}

func TestRedirect(t *testing.T) {
	Convey("should return the correct redirect value", t, func() {
		r := route.New(route.Options{Path: "/test"})
		So(r.Redirect(), ShouldEqual, false)
	})
}

func TestByRoutePath(t *testing.T) {
	Convey("should sort the given routes by path", t, func() {
		rl := []route.Route{*route.New(route.Options{Path: "/test"}), *route.New(route.Options{Path: "/"})}
		sort.Sort(route.ByRoutePath(rl))
		So(rl[0].Path(), ShouldEqual, "/")
		So(rl[1].Path(), ShouldEqual, "/test")
	})
}

func TestListByMux(t *testing.T) {
	Convey("should return the list of routes by the given mux", t, func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/test", func(http.ResponseWriter, *http.Request) {})
		mux.HandleFunc("/", func(http.ResponseWriter, *http.Request) {})
		rl := route.ListByMux(mux)
		sort.Sort(route.ByRoutePath(rl))
		So(rl[0].Path(), ShouldEqual, "")
		So(rl[0].Explicit(), ShouldEqual, false)
		So(rl[0].Pattern(), ShouldEqual, "/")
		So(rl[0].Redirect(), ShouldEqual, true)
		So(rl[1].Path(), ShouldEqual, "/")
		So(rl[1].Pattern(), ShouldEqual, "/")
		So(rl[1].Explicit(), ShouldEqual, true)
		So(rl[1].Redirect(), ShouldEqual, false)
		So(rl[2].Path(), ShouldEqual, "/test")
		So(rl[2].Pattern(), ShouldEqual, "/test")
		So(rl[2].Explicit(), ShouldEqual, true)
		So(rl[2].Redirect(), ShouldEqual, false)
	})
}
