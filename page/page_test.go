/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package page_test

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/devfacet/goweb/page"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("should return a new page", t, func() {
		p, err := page.New(page.Options{URLPath: "/test"})
		So(err, ShouldBeNil)
		So(p.URLPath(), ShouldEqual, "/test")
		So(p.MatchAll(), ShouldEqual, false)
	})

	Convey("should fail to return a new page due to invalid page url", t, func() {
		p, err := page.New(page.Options{})
		So(err, ShouldBeError, errors.New("invalid url path"))
		So(p, ShouldBeNil)
	})

	Convey("should return a new page with the given file and file system", t, func() {
		fs := http.FileSystem(http.Dir("./"))
		p, err := page.New(page.Options{URLPath: "/test", FilePath: "/test.html", FileSystem: &fs})
		So(err, ShouldBeNil)
		So(p, ShouldNotBeNil)
	})

	Convey("should fail to return a new page due to invalid file path", t, func() {
		fs := http.FileSystem(http.Dir("./"))
		p, err := page.New(page.Options{URLPath: "/test", FilePath: "error.html", FileSystem: &fs})
		So(err, ShouldBeError, errors.New("failed to open file due to open error.html: no such file or directory"))
		So(p, ShouldBeNil)
	})

	Convey("should fail to return a new page due to invalid file", t, func() {
		fs := http.FileSystem(http.Dir("./"))
		p, err := page.New(page.Options{URLPath: "/test", FilePath: ".", FileSystem: &fs})
		So(err, ShouldBeError, errors.New("failed to read file due to read .: is a directory"))
		So(p, ShouldBeNil)
	})

	Convey("should fail to return a new page due to missing file system", t, func() {
		p, err := page.New(page.Options{URLPath: "/test", FilePath: "error.html"})
		So(err, ShouldBeError, errors.New("invalid file system"))
		So(p, ShouldBeNil)
	})

	Convey("should fail to return a new page due to invalid page content", t, func() {
		p, err := page.New(page.Options{
			URLPath:      "/test",
			Content:      "{{.Test",
			TemplateData: struct{ Test string }{Test: "test"},
		})
		So(err, ShouldBeError, errors.New("failed to parse template due to template: /test:1: unclosed action"))
		So(p, ShouldBeNil)
	})
}

func TestURLPath(t *testing.T) {
	Convey("should return the given url path", t, func() {
		pages := []struct {
			in  string
			out string
		}{
			{"test", "test"},
			{"/test", "/test"},
			{"test/", "test/"},
			{"/test/", "/test/"},
		}
		for _, v := range pages {
			p, err := page.New(page.Options{URLPath: v.in})
			So(err, ShouldBeNil)
			So(p.URLPath(), ShouldEqual, v.out)
		}
	})
}

func TestMatchAll(t *testing.T) {
	Convey("should return the correct match all value", t, func() {
		p1, err := page.New(page.Options{URLPath: "/test", MatchAll: true})
		So(err, ShouldBeNil)
		So(p1.MatchAll(), ShouldEqual, true)

		p2, err := page.New(page.Options{URLPath: "/test"})
		So(err, ShouldBeNil)
		So(p2.MatchAll(), ShouldEqual, false)
	})
}

func TestTemplateExecute(t *testing.T) {
	Convey("should execute page template", t, func() {
		fs := http.FileSystem(http.Dir("./"))

		p1, err := page.New(page.Options{URLPath: "/test"})
		So(err, ShouldBeNil)
		So(p1.TemplateExecute(ioutil.Discard, nil), ShouldBeNil)

		p2, err := page.New(page.Options{URLPath: "/test", FilePath: "/test.html", FileSystem: &fs})
		So(err, ShouldBeNil)
		So(p2.TemplateExecute(ioutil.Discard, nil), ShouldBeNil)

		p3, err := page.New(page.Options{URLPath: "/test", Content: "{{.Test}}", TemplateData: struct{ Test string }{Test: "test"}})
		So(err, ShouldBeNil)
		b1 := bytes.Buffer{}
		w1 := bufio.NewWriter(&b1)
		So(p3.TemplateExecute(w1, nil), ShouldBeNil)
		w1.Flush()
		So(b1.String(), ShouldEqual, "test")

		p4, err := page.New(page.Options{URLPath: "/test", FilePath: "/test.html", FileSystem: &fs, TemplateData: struct{ Test string }{Test: "foo"}})
		So(err, ShouldBeNil)
		b2 := bytes.Buffer{}
		w2 := bufio.NewWriter(&b2)
		So(p4.TemplateExecute(w2, nil), ShouldBeNil)
		w2.Flush()
		So(b2.String(), ShouldEqual, "test foo")
	})

	Convey("should fail to execute page template", t, func() {
		p1, err := page.New(page.Options{URLPath: "/test", Content: "{{.Test}}", TemplateData: struct{ test string }{test: "test"}})
		So(err, ShouldBeNil)
		So(p1.TemplateExecute(ioutil.Discard, nil), ShouldBeError, errors.New(`failed to execute template due to template: /test:1:2: executing "/test" at <.Test>: can't evaluate field Test in type struct { test string }`))
	})
}
