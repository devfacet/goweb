/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package server_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/devfacet/goweb/page"
	"github.com/devfacet/goweb/server"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("should create a new server", t, func() {
		s := server.New(server.Options{})
		So(s.ID(), ShouldNotBeEmpty)
		So(s.Address(), ShouldNotBeEmpty)
	})
}

func TestListenAll(t *testing.T) {
	Convey("should listen all", t, func() {
		servers := []*server.Server{
			server.New(server.Options{ID: "web", Address: "localhost:3061"}),
			server.New(server.Options{ID: "api", Address: "localhost:3062"}),
		}
		p, err := page.New(page.Options{URLPath: "/foo", Content: "foo"})
		So(err, ShouldBeNil)
		servers[0].AddPage(p)
		p, err = page.New(page.Options{URLPath: "/bar", Content: "bar"})
		So(err, ShouldBeNil)
		servers[1].AddPage(p)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Done()
			server.ListenAll(servers...)
		}()
		wg.Wait()

		resp, err := http.Get(fmt.Sprintf("http://%s/foo", servers[0].Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		So(string(b), ShouldEqual, "foo")

		resp, err = http.Get(fmt.Sprintf("http://%s/bar", servers[1].Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ = ioutil.ReadAll(resp.Body)
		So(string(b), ShouldEqual, "bar")

		servers[0].Close()
		servers[1].Close()
	})

	Convey("should fail to listen all", t, func() {
		servers := []*server.Server{
			server.New(server.Options{ID: "web", Address: "localhost:3060"}),
			server.New(server.Options{ID: "api", Address: "localhost:3060"}),
		}
		p, err := page.New(page.Options{URLPath: "/foo", Content: "foo"})
		So(err, ShouldBeNil)
		servers[0].AddPage(p)
		p, err = page.New(page.Options{URLPath: "/bar", Content: "bar"})
		So(err, ShouldBeNil)
		servers[1].AddPage(p)

		if err := server.ListenAll(servers...); err != nil {
			So(err, ShouldBeError, "listen tcp 127.0.0.1:3060: bind: address already in use")
		}

		servers[0].Close()
		servers[1].Close()
	})
}

func TestID(t *testing.T) {
	Convey("should return right id", t, func() {
		s := server.New(server.Options{ID: "test"})
		So(s.ID(), ShouldEqual, "test")
	})
}

func TestAddress(t *testing.T) {
	Convey("should return right address", t, func() {
		s := server.New(server.Options{Address: "localhost:3000"})
		So(s.Address(), ShouldEqual, "localhost:3000")
	})
}

func TestPathRoot(t *testing.T) {
	Convey("should return right root path", t, func() {
		s := server.New(server.Options{})
		So(s.PathRoot(), ShouldEqual, "")

		s = server.New(server.Options{PathPrefix: "test"})
		So(s.PathRoot(), ShouldEqual, "/test/")

		s = server.New(server.Options{PathPrefix: "/test"})
		So(s.PathRoot(), ShouldEqual, "/test/")

		s = server.New(server.Options{PathPrefix: "test/"})
		So(s.PathRoot(), ShouldEqual, "/test/")

		s = server.New(server.Options{PathPrefix: "/test/"})
		So(s.PathRoot(), ShouldEqual, "/test/")
	})
}

func TestListen(t *testing.T) {
	Convey("should listen", t, func() {
		s := server.New(server.Options{})
		p, err := page.New(page.Options{URLPath: "/test", Content: "test"})
		So(err, ShouldBeNil)
		s.AddPage(p)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Done()
			s.Listen()
		}()
		wg.Wait()

		resp, err := http.Get(fmt.Sprintf("http://%s/test", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		So(string(b), ShouldEqual, "test")

		s.Close()
	})

	Convey("should fail to listen due to invalid address", t, func() {
		s := server.New(server.Options{Address: "localhost"})
		So(s.Listen(), ShouldBeError, "listen tcp: address localhost: missing port in address")
	})
}

func TestClose(t *testing.T) {
	Convey("should close listeners and connections", t, func() {
		s := server.New(server.Options{})
		So(s.Close(), ShouldBeNil)
		go func() {
			s.Listen()
		}()
		select {
		case <-time.After(time.Millisecond):
			So(s.Close(), ShouldBeNil)
		}
	})
}

func TestShutdown(t *testing.T) {
	Convey("should close listeners and connections", t, func() {
		s := server.New(server.Options{})
		So(s.Shutdown(), ShouldBeNil)
		go func() {
			s.Listen()
		}()
		select {
		case <-time.After(time.Millisecond):
			So(s.Shutdown(), ShouldBeNil)
		}
	})
}

func TestRoutes(t *testing.T) {
	Convey("should return list of routes", t, func() {
		s := server.New(server.Options{})
		s.AddHandler("/", *new(http.Handler))
		s.AddHandlerFunc("/test", func(http.ResponseWriter, *http.Request) {})
		s.AddHandlerFunc("test/", func(http.ResponseWriter, *http.Request) {})
		rl := s.Routes()
		So(rl[0].Path(), ShouldEqual, "")
		So(rl[0].Pattern(), ShouldEqual, "/")
		So(rl[0].Explicit(), ShouldEqual, false)
		So(rl[0].Redirect(), ShouldEqual, true)
		So(rl[1].Path(), ShouldEqual, "/")
		So(rl[1].Pattern(), ShouldEqual, "/")
		So(rl[1].Explicit(), ShouldEqual, true)
		So(rl[1].Redirect(), ShouldEqual, false)
		So(rl[2].Path(), ShouldEqual, "/test")
		So(rl[2].Pattern(), ShouldEqual, "/test")
		So(rl[2].Explicit(), ShouldEqual, true)
		So(rl[2].Redirect(), ShouldEqual, false)
		So(rl[3].Path(), ShouldEqual, "/test/")
		So(rl[3].Pattern(), ShouldEqual, "/test/")
		So(rl[3].Explicit(), ShouldEqual, true)
		So(rl[3].Redirect(), ShouldEqual, false)
	})
}

func TestAddHandler(t *testing.T) {
	Convey("should add a handler", t, func() {
		s := server.New(server.Options{})
		s.AddHandlerFunc("/", func(http.ResponseWriter, *http.Request) {})
		s.AddHandlerFunc("/test", func(http.ResponseWriter, *http.Request) {})
		s.AddHandlerFunc("test/", func(http.ResponseWriter, *http.Request) {})
		rl := s.Routes()
		So(rl[0].Path(), ShouldEqual, "")
		So(rl[0].Pattern(), ShouldEqual, "/")
		So(rl[0].Explicit(), ShouldEqual, false)
		So(rl[0].Redirect(), ShouldEqual, true)
		So(rl[1].Path(), ShouldEqual, "/")
		So(rl[1].Pattern(), ShouldEqual, "/")
		So(rl[1].Explicit(), ShouldEqual, true)
		So(rl[1].Redirect(), ShouldEqual, false)
		So(rl[2].Path(), ShouldEqual, "/test")
		So(rl[2].Pattern(), ShouldEqual, "/test")
		So(rl[2].Explicit(), ShouldEqual, true)
		So(rl[2].Redirect(), ShouldEqual, false)
		So(rl[3].Path(), ShouldEqual, "/test/")
		So(rl[3].Pattern(), ShouldEqual, "/test/")
		So(rl[3].Explicit(), ShouldEqual, true)
		So(rl[3].Redirect(), ShouldEqual, false)
	})

	Convey("should add a handler (prefix)", t, func() {
		s := server.New(server.Options{PathPrefix: "foo"})
		s.AddHandler("/", *new(http.Handler))
		s.AddHandler("/test", *new(http.Handler))
		s.AddHandler("test/", *new(http.Handler))
		rl := s.Routes()
		So(rl[0].Path(), ShouldEqual, "/foo")
		So(rl[0].Pattern(), ShouldEqual, "/foo/")
		So(rl[0].Explicit(), ShouldEqual, false)
		So(rl[0].Redirect(), ShouldEqual, true)
		So(rl[1].Path(), ShouldEqual, "/foo/")
		So(rl[1].Pattern(), ShouldEqual, "/foo/")
		So(rl[1].Explicit(), ShouldEqual, true)
		So(rl[1].Redirect(), ShouldEqual, false)
		So(rl[2].Path(), ShouldEqual, "/foo/test")
		So(rl[2].Pattern(), ShouldEqual, "/foo/test")
		So(rl[2].Explicit(), ShouldEqual, true)
		So(rl[2].Redirect(), ShouldEqual, false)
		So(rl[3].Path(), ShouldEqual, "/foo/test/")
		So(rl[3].Pattern(), ShouldEqual, "/foo/test/")
		So(rl[3].Explicit(), ShouldEqual, true)
		So(rl[3].Redirect(), ShouldEqual, false)

		s = server.New(server.Options{PathPrefix: "/foo"})
		s.AddHandler("/", *new(http.Handler))
		s.AddHandler("/test", *new(http.Handler))
		s.AddHandler("test/", *new(http.Handler))
		rl = s.Routes()
		So(rl[0].Path(), ShouldEqual, "/foo")
		So(rl[0].Pattern(), ShouldEqual, "/foo/")
		So(rl[0].Explicit(), ShouldEqual, false)
		So(rl[0].Redirect(), ShouldEqual, true)
		So(rl[1].Path(), ShouldEqual, "/foo/")
		So(rl[1].Pattern(), ShouldEqual, "/foo/")
		So(rl[1].Explicit(), ShouldEqual, true)
		So(rl[1].Redirect(), ShouldEqual, false)
		So(rl[2].Path(), ShouldEqual, "/foo/test")
		So(rl[2].Pattern(), ShouldEqual, "/foo/test")
		So(rl[2].Explicit(), ShouldEqual, true)
		So(rl[2].Redirect(), ShouldEqual, false)
		So(rl[3].Path(), ShouldEqual, "/foo/test/")
		So(rl[3].Pattern(), ShouldEqual, "/foo/test/")
		So(rl[3].Explicit(), ShouldEqual, true)
		So(rl[3].Redirect(), ShouldEqual, false)

		s = server.New(server.Options{PathPrefix: "/foo/"})
		s.AddHandler("/", *new(http.Handler))
		s.AddHandler("/test", *new(http.Handler))
		s.AddHandler("test/", *new(http.Handler))
		rl = s.Routes()
		So(rl[0].Path(), ShouldEqual, "/foo")
		So(rl[0].Pattern(), ShouldEqual, "/foo/")
		So(rl[0].Explicit(), ShouldEqual, false)
		So(rl[0].Redirect(), ShouldEqual, true)
		So(rl[1].Path(), ShouldEqual, "/foo/")
		So(rl[1].Pattern(), ShouldEqual, "/foo/")
		So(rl[1].Explicit(), ShouldEqual, true)
		So(rl[1].Redirect(), ShouldEqual, false)
		So(rl[2].Path(), ShouldEqual, "/foo/test")
		So(rl[2].Pattern(), ShouldEqual, "/foo/test")
		So(rl[2].Explicit(), ShouldEqual, true)
		So(rl[2].Redirect(), ShouldEqual, false)
		So(rl[3].Path(), ShouldEqual, "/foo/test/")
		So(rl[3].Pattern(), ShouldEqual, "/foo/test/")
		So(rl[3].Explicit(), ShouldEqual, true)
		So(rl[3].Redirect(), ShouldEqual, false)
	})
}

func TestAddHandlerFunc(t *testing.T) {
	Convey("should add a handler by handler function", t, func() {
		s := server.New(server.Options{})
		s.AddHandlerFunc("/", func(http.ResponseWriter, *http.Request) {})
		s.AddHandlerFunc("/test", func(http.ResponseWriter, *http.Request) {})
		s.AddHandlerFunc("test/", func(http.ResponseWriter, *http.Request) {})
		rl := s.Routes()
		So(rl[0].Path(), ShouldEqual, "")
		So(rl[0].Pattern(), ShouldEqual, "/")
		So(rl[0].Explicit(), ShouldEqual, false)
		So(rl[0].Redirect(), ShouldEqual, true)
		So(rl[1].Path(), ShouldEqual, "/")
		So(rl[1].Pattern(), ShouldEqual, "/")
		So(rl[1].Explicit(), ShouldEqual, true)
		So(rl[1].Redirect(), ShouldEqual, false)
		So(rl[2].Path(), ShouldEqual, "/test")
		So(rl[2].Pattern(), ShouldEqual, "/test")
		So(rl[2].Explicit(), ShouldEqual, true)
		So(rl[2].Redirect(), ShouldEqual, false)
		So(rl[3].Path(), ShouldEqual, "/test/")
		So(rl[3].Pattern(), ShouldEqual, "/test/")
		So(rl[3].Explicit(), ShouldEqual, true)
		So(rl[3].Redirect(), ShouldEqual, false)
	})

	Convey("should add a handler by handler function (prefix)", t, func() {
		s := server.New(server.Options{PathPrefix: "foo"})
		s.AddHandlerFunc("/", func(http.ResponseWriter, *http.Request) {})
		s.AddHandlerFunc("/test", func(http.ResponseWriter, *http.Request) {})
		s.AddHandlerFunc("test/", func(http.ResponseWriter, *http.Request) {})
		rl := s.Routes()
		So(rl[0].Path(), ShouldEqual, "/foo")
		So(rl[0].Pattern(), ShouldEqual, "/foo/")
		So(rl[0].Explicit(), ShouldEqual, false)
		So(rl[0].Redirect(), ShouldEqual, true)
		So(rl[1].Path(), ShouldEqual, "/foo/")
		So(rl[1].Pattern(), ShouldEqual, "/foo/")
		So(rl[1].Explicit(), ShouldEqual, true)
		So(rl[1].Redirect(), ShouldEqual, false)
		So(rl[2].Path(), ShouldEqual, "/foo/test")
		So(rl[2].Pattern(), ShouldEqual, "/foo/test")
		So(rl[2].Explicit(), ShouldEqual, true)
		So(rl[2].Redirect(), ShouldEqual, false)
		So(rl[3].Path(), ShouldEqual, "/foo/test/")
		So(rl[3].Pattern(), ShouldEqual, "/foo/test/")
		So(rl[3].Explicit(), ShouldEqual, true)
		So(rl[3].Redirect(), ShouldEqual, false)

		s = server.New(server.Options{PathPrefix: "/foo"})
		s.AddHandlerFunc("/", func(http.ResponseWriter, *http.Request) {})
		s.AddHandlerFunc("/test", func(http.ResponseWriter, *http.Request) {})
		s.AddHandlerFunc("test/", func(http.ResponseWriter, *http.Request) {})
		rl = s.Routes()
		So(rl[0].Path(), ShouldEqual, "/foo")
		So(rl[0].Pattern(), ShouldEqual, "/foo/")
		So(rl[0].Explicit(), ShouldEqual, false)
		So(rl[0].Redirect(), ShouldEqual, true)
		So(rl[1].Path(), ShouldEqual, "/foo/")
		So(rl[1].Pattern(), ShouldEqual, "/foo/")
		So(rl[1].Explicit(), ShouldEqual, true)
		So(rl[1].Redirect(), ShouldEqual, false)
		So(rl[2].Path(), ShouldEqual, "/foo/test")
		So(rl[2].Pattern(), ShouldEqual, "/foo/test")
		So(rl[2].Explicit(), ShouldEqual, true)
		So(rl[2].Redirect(), ShouldEqual, false)
		So(rl[3].Path(), ShouldEqual, "/foo/test/")
		So(rl[3].Pattern(), ShouldEqual, "/foo/test/")
		So(rl[3].Explicit(), ShouldEqual, true)
		So(rl[3].Redirect(), ShouldEqual, false)

		s = server.New(server.Options{PathPrefix: "/foo/"})
		s.AddHandlerFunc("/", func(http.ResponseWriter, *http.Request) {})
		s.AddHandlerFunc("/test", func(http.ResponseWriter, *http.Request) {})
		s.AddHandlerFunc("test/", func(http.ResponseWriter, *http.Request) {})
		rl = s.Routes()
		So(rl[0].Path(), ShouldEqual, "/foo")
		So(rl[0].Pattern(), ShouldEqual, "/foo/")
		So(rl[0].Explicit(), ShouldEqual, false)
		So(rl[0].Redirect(), ShouldEqual, true)
		So(rl[1].Path(), ShouldEqual, "/foo/")
		So(rl[1].Pattern(), ShouldEqual, "/foo/")
		So(rl[1].Explicit(), ShouldEqual, true)
		So(rl[1].Redirect(), ShouldEqual, false)
		So(rl[2].Path(), ShouldEqual, "/foo/test")
		So(rl[2].Pattern(), ShouldEqual, "/foo/test")
		So(rl[2].Explicit(), ShouldEqual, true)
		So(rl[2].Redirect(), ShouldEqual, false)
		So(rl[3].Path(), ShouldEqual, "/foo/test/")
		So(rl[3].Pattern(), ShouldEqual, "/foo/test/")
		So(rl[3].Explicit(), ShouldEqual, true)
		So(rl[3].Redirect(), ShouldEqual, false)
	})
}

func TestAddPage(t *testing.T) {
	Convey("should add a page", t, func() {
		s := server.New(server.Options{})
		p, err := page.New(page.Options{
			URLPath:      "/foo",
			Content:      "{{.Body}}",
			TemplateData: struct{ Body string }{Body: "foo"},
		})
		So(err, ShouldBeNil)
		So(s.AddPage(p), ShouldBeNil)

		s = server.New(server.Options{PathPrefix: "foo"})
		p, err = page.New(page.Options{
			URLPath:      "/foo",
			Content:      "{{.Body}}",
			TemplateData: struct{ Body string }{Body: "foo"},
		})
		So(err, ShouldBeNil)
		So(s.AddPage(p), ShouldBeNil)
	})

	Convey("should fail to add a page", t, func() {
		s := server.New(server.Options{})
		p := &page.Page{}
		So(s.AddPage(p), ShouldBeError, errors.New("invalid page url"))
	})

	Convey("should serve the right page", t, func() {
		s := server.New(server.Options{})
		p, err := page.New(page.Options{
			URLPath:      "/foo",
			Content:      "{{.Body}}",
			TemplateData: struct{ Body string }{Body: "foo"},
		})
		So(err, ShouldBeNil)
		So(s.AddPage(p), ShouldBeNil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Done()
			s.Listen()
		}()
		wg.Wait()

		resp, err := http.Get(fmt.Sprintf("http://%s/foo", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		So(string(b), ShouldEqual, "foo")

		resp, err = http.Get(fmt.Sprintf("http://%s/bar", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ = ioutil.ReadAll(resp.Body)
		So(resp.StatusCode, ShouldEqual, 404)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
		So(strings.TrimSpace(string(b)), ShouldEqual, "404 page not found")

		s.Close()
	})

	Convey("should serve the right page with match all", t, func() {
		s := server.New(server.Options{})
		p, err := page.New(page.Options{
			URLPath:      "/foo/",
			MatchAll:     true,
			Content:      "{{.Body}}",
			TemplateData: struct{ Body string }{Body: "foo"},
		})
		So(err, ShouldBeNil)
		So(s.AddPage(p), ShouldBeNil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Done()
			s.Listen()
		}()
		wg.Wait()

		resp, err := http.Get(fmt.Sprintf("http://%s/foo", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		So(string(b), ShouldEqual, "foo")

		resp, err = http.Get(fmt.Sprintf("http://%s/foo/", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ = ioutil.ReadAll(resp.Body)
		So(string(b), ShouldEqual, "foo")

		resp, err = http.Get(fmt.Sprintf("http://%s/foo/bar", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ = ioutil.ReadAll(resp.Body)
		So(string(b), ShouldEqual, "foo")

		resp, err = http.Get(fmt.Sprintf("http://%s/foo/bar/baz", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ = ioutil.ReadAll(resp.Body)
		So(string(b), ShouldEqual, "foo")

		s.Close()

		s = server.New(server.Options{})
		p, err = page.New(page.Options{
			URLPath:      "/foo",
			MatchAll:     false,
			Content:      "{{.Body}}",
			TemplateData: struct{ Body string }{Body: "foo"},
		})
		So(err, ShouldBeNil)
		So(s.AddPage(p), ShouldBeNil)

		wg = sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Done()
			s.Listen()
		}()
		wg.Wait()

		resp, err = http.Get(fmt.Sprintf("http://%s/foo", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ = ioutil.ReadAll(resp.Body)
		So(string(b), ShouldEqual, "foo")

		resp, err = http.Get(fmt.Sprintf("http://%s/foo/", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ = ioutil.ReadAll(resp.Body)
		So(resp.StatusCode, ShouldEqual, 404)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
		So(strings.TrimSpace(string(b)), ShouldEqual, "404 page not found")

		resp, err = http.Get(fmt.Sprintf("http://%s/foo/bar", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ = ioutil.ReadAll(resp.Body)
		So(resp.StatusCode, ShouldEqual, 404)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
		So(strings.TrimSpace(string(b)), ShouldEqual, "404 page not found")

		s.Close()
	})

	Convey("should serve the home page)", t, func() {
		s := server.New(server.Options{})
		p, err := page.New(page.Options{
			URLPath:      "/",
			Content:      "{{.Body}}",
			TemplateData: struct{ Body string }{Body: "home"},
		})
		So(err, ShouldBeNil)
		So(s.AddPage(p), ShouldBeNil)
		p, err = page.New(page.Options{
			URLPath:      "/foo",
			Content:      "{{.Body}}",
			TemplateData: struct{ Body string }{Body: "foo"},
		})
		So(err, ShouldBeNil)
		So(s.AddPage(p), ShouldBeNil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Done()
			s.Listen()
		}()
		wg.Wait()

		resp, err := http.Get(fmt.Sprintf("http://%s/", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		So(string(b), ShouldEqual, "home")

		resp, err = http.Get(fmt.Sprintf("http://%s/foo", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ = ioutil.ReadAll(resp.Body)
		So(string(b), ShouldEqual, "foo")

		resp, err = http.Get(fmt.Sprintf("http://%s/bar", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ = ioutil.ReadAll(resp.Body)
		So(resp.StatusCode, ShouldEqual, 404)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(strings.TrimSpace(string(b)), ShouldEqual, `{"statusCode":404,"message":"Not Found"}`)

		s.Close()
	})

	Convey("should fail to serve a page due to template data issue", t, func() {
		s := server.New(server.Options{})
		p, err := page.New(page.Options{
			URLPath:      "/foo",
			Content:      "{{.Body}}",
			TemplateData: struct{ Foo string }{Foo: "foo"},
		})
		So(err, ShouldBeNil)
		So(s.AddPage(p), ShouldBeNil)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Done()
			s.Listen()
		}()
		wg.Wait()

		resp, err := http.Get(fmt.Sprintf("http://%s/foo", s.Address()))
		So(err, ShouldBeNil)
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		So(resp.StatusCode, ShouldEqual, 500)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(strings.TrimSpace(string(b)), ShouldEqual, `{"statusCode":500,"message":"Internal Server Error"}`)

		s.Close()
	})
}

func TestAddPages(t *testing.T) {
	Convey("should add pages", t, func() {
		s := server.New(server.Options{})
		p1, err := page.New(page.Options{
			URLPath:      "/foo",
			Content:      "{{.Body}}",
			TemplateData: struct{ Body string }{Body: "foo"},
		})
		So(err, ShouldBeNil)
		p2, err := page.New(page.Options{
			URLPath:      "/bar",
			Content:      "{{.Body}}",
			TemplateData: struct{ Body string }{Body: "bar"},
		})
		So(err, ShouldBeNil)
		pages := []*page.Page{p1, p2}
		So(s.AddPages(pages...), ShouldBeNil)
	})

	Convey("should fail to add pages", t, func() {
		s := server.New(server.Options{})
		pages := []*page.Page{
			&page.Page{},
		}
		So(s.AddPages(pages...), ShouldBeError, errors.New("invalid page url"))
	})
}
