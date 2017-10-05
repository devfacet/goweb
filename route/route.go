/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package route provides functions for HTTP routes
package route

import (
	"net/http"
	"reflect"
	"sort"
	"strings"
)

// Options represents the options than can be set when creating a new route
type Options struct {
	// Path holds path value
	Path string
}

// New returns a route by the given options
func New(o Options) *Route {
	// Init the route
	route := Route{
		isInit:   true,
		path:     o.Path,
		pattern:  o.Path,
		explicit: true,
		redirect: false,
	}

	return &route
}

// Route represents an HTTP route
type Route struct {
	isInit   bool
	path     string
	pattern  string
	explicit bool
	redirect bool
}

// Path returns the route path
func (route *Route) Path() string {
	return route.path
}

// Pattern returns the pattern value
func (route *Route) Pattern() string {
	return route.pattern
}

// Explicit returns the explicit value
func (route *Route) Explicit() bool {
	return route.explicit
}

// Redirect returns the redirect value
func (route *Route) Redirect() bool {
	return route.redirect
}

// ByRoutePath implements sort.Interface for []Route
type ByRoutePath []Route

func (r ByRoutePath) Len() int           { return len(r) }
func (r ByRoutePath) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ByRoutePath) Less(i, j int) bool { return r[i].Path() < r[j].Path() }

// ListByMux returns the route list by the given ServeMux
func ListByMux(mux *http.ServeMux) []Route {
	result := []Route{}

	// Iterate over routes
	e := reflect.ValueOf(mux).Elem().FieldByName("m")
	for _, v := range e.MapKeys() {
		//fmt.Printf("%s > %#v\n", v, e.MapIndex(v)) // for debug
		rh := false
		h := e.MapIndex(v).FieldByName("h").Elem()
		if strings.Contains(reflect.Indirect(h).String(), "http.redirectHandler") {
			rh = true
		} else {
			rh = false
		}
		r := Route{
			path:     v.String(),
			pattern:  e.MapIndex(v).FieldByName("pattern").String(),
			explicit: e.MapIndex(v).FieldByName("explicit").Bool(),
			redirect: rh,
		}
		result = append(result, r)
	}

	sort.Sort(ByRoutePath(result))

	return result
}
