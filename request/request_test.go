/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package request_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"os"

	"github.com/devfacet/goweb/request"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("should return a new request", t, func() {
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil)})
		So(req.ContentType(), ShouldBeEmpty)

		req = request.New(request.Options{Writer: *new(http.ResponseWriter)})
		So(req.ContentType(), ShouldBeEmpty)

		req = request.New(request.Options{})
		So(req.ContentType(), ShouldBeEmpty)
	})
}

func TestContentType(t *testing.T) {
	Convey("should return the given content type", t, func() {
		r := httptest.NewRequest("GET", "http://localhost", nil)
		r.Header.Set("Content-Type", "application/json")
		req := request.New(request.Options{Request: r})

		So(req.ContentType(), ShouldEqual, "application/json")
	})
}

func TestReply(t *testing.T) {
	Convey("should reply with empty response", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		req.Reply(nil)
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
		So(string(b), ShouldBeEmpty)
	})

	Convey("should reply with the given []byte value", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		req.Reply([]byte("test"))
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
		So(string(b), ShouldEqual, "test")
	})

	Convey("should reply with the given string value", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		req.Reply("test")
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
		So(string(b), ShouldEqual, "test")
	})

	Convey("should reply with the given int value", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		var v int = 1
		req.Reply(v)
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
		So(string(b), ShouldEqual, "1")
	})

	Convey("should reply with the given int64 value", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		var v int64 = 1
		req.Reply(v)
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
		So(string(b), ShouldEqual, "1")
	})

	Convey("should reply with the given float64 value", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		var v float64 = 3.14159265359
		req.Reply(v)
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
		So(string(b), ShouldEqual, "3.14159265359")
	})

	Convey("should reply with default success", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		req.Reply(request.Success{})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(string(b), ShouldEqual, `{"statusCode":200,"message":"OK"}`)
	})

	Convey("should reply with the given success code", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		req.Reply(request.Success{StatusCode: 200})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(string(b), ShouldEqual, `{"statusCode":200,"message":"OK"}`)
	})

	Convey("should reply with the given success message", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		req.Reply(request.Success{StatusCode: 200, Message: "done"})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(string(b), ShouldEqual, `{"statusCode":200,"message":"done"}`)
	})

	Convey("should reply with the given success code and callback", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost/?callback=cb", nil), Writer: w})
		req.Reply(request.Success{StatusCode: 200})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/javascript")
		So(string(b), ShouldEqual, `cb({"statusCode":200,"message":"OK"})`)
	})

	Convey("should reply with default error", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		req.Reply(request.Error{})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 500)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(string(b), ShouldEqual, `{"statusCode":500,"message":"Internal Server Error"}`)
	})

	Convey("should reply with the given error code", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		req.Reply(request.Error{StatusCode: 400})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 400)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(string(b), ShouldEqual, `{"statusCode":400,"message":"Bad Request"}`)
	})

	Convey("should reply with the given error message", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		req.Reply(request.Error{StatusCode: 400, Message: "test"})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 400)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(string(b), ShouldEqual, `{"statusCode":400,"message":"test"}`)
	})

	Convey("should reply with the given error code and callback", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost/?callback=cb", nil), Writer: w})
		req.Reply(request.Error{StatusCode: 400})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 400)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/javascript")
		So(string(b), ShouldEqual, `cb({"statusCode":400,"message":"Bad Request"})`)
	})

	Convey("should reply with default error (chan)", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		v := make(chan int)
		req.Reply(request.Error{Error: v})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 500)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(string(b), ShouldEqual, `{"statusCode":500,"message":"Internal Server Error"}`)
	})

	Convey("should reply with default error and callback (chan)", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost/?callback=cb", nil), Writer: w})
		v := make(chan int)
		req.Reply(request.Error{Error: v})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 500)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/javascript")
		So(string(b), ShouldEqual, `cb({"statusCode":500,"message":"Internal Server Error"})`)
	})

	Convey("should reply with the given empty struct", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		req.Reply(struct{}{})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(string(b), ShouldEqual, `{}`)
	})

	Convey("should reply with the given custom struct", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		req.Reply(struct{ Test string }{Test: "test"})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(string(b), ShouldEqual, `{"Test":"test"}`)
	})

	Convey("should fail to reply with the given custom struct", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost", nil), Writer: w})
		v := make(chan int)
		req.Reply(struct{ Test chan int }{Test: v})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 500)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/json")
		So(string(b), ShouldEqual, `{"statusCode":500,"message":"Internal Server Error"}`)
	})

	Convey("should reply with the given custom struct and callback", t, func() {
		w := httptest.NewRecorder()
		req := request.New(request.Options{Request: httptest.NewRequest("GET", "http://localhost/?callback=cb", nil), Writer: w})
		req.Reply(struct{ Test string }{Test: "test"})
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "application/javascript")
		So(string(b), ShouldEqual, `cb({"Test":"test"})`)
	})

	Convey("should reply with the given unknown type value", t, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://localhost", nil)
		req := request.New(request.Options{Request: r, Writer: w})
		v := []string{"test"}
		req.Reply(v)
		resp := w.Result()
		b, _ := ioutil.ReadAll(resp.Body)

		So(resp.StatusCode, ShouldEqual, 200)
		So(resp.Header.Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
		So(string(b), ShouldEqual, "[test]")
	})
}

func TestFormFiles(t *testing.T) {
	Convey("should return a form file", t, func() {
		f, err := os.Open("test.txt")
		defer f.Close()
		So(err, ShouldBeNil)

		buf := &bytes.Buffer{}
		mw := multipart.NewWriter(buf)
		w, err := mw.CreateFormFile("file", "test.txt")
		So(err, ShouldBeNil)
		_, err = io.Copy(w, f)
		So(err, ShouldBeNil)

		contentType := mw.FormDataContentType()
		mw.Close()

		r := httptest.NewRequest("POST", "http://localhost", buf)
		r.Header.Set("Content-Type", contentType)
		req := request.New(request.Options{Request: r})
		ff := req.FormFiles()

		So(req.ContentType(), ShouldEqual, "multipart/form-data")
		So(len(ff), ShouldEqual, 1)
		So(ff[0].FieldName(), ShouldEqual, "file")
		So(ff[0].FileName(), ShouldEqual, "test.txt")
		So(ff[0].ContentType(), ShouldEqual, "application/octet-stream")
		So(ff[0].Multiple(), ShouldEqual, false)
		So(ff[0].FileHeader(), ShouldNotBeNil)
	})

	Convey("should return multiple form files", t, func() {
		f, err := os.Open("test.txt")
		defer f.Close()
		So(err, ShouldBeNil)

		buf := &bytes.Buffer{}
		mw := multipart.NewWriter(buf)
		w, err := mw.CreateFormFile("file", "test1.txt")
		So(err, ShouldBeNil)
		_, err = io.Copy(w, f)
		So(err, ShouldBeNil)

		w, err = mw.CreateFormFile("file", "test2.txt")
		So(err, ShouldBeNil)
		_, err = io.Copy(w, f)
		So(err, ShouldBeNil)

		contentType := mw.FormDataContentType()
		mw.Close()

		r := httptest.NewRequest("POST", "http://localhost", buf)
		r.Header.Set("Content-Type", contentType)
		req := request.New(request.Options{Request: r})
		ff := req.FormFiles()

		So(req.ContentType(), ShouldEqual, "multipart/form-data")
		So(len(ff), ShouldEqual, 2)
		So(ff[0].FieldName(), ShouldEqual, "file")
		So(ff[0].FileName(), ShouldEqual, "test1.txt")
		So(ff[0].ContentType(), ShouldEqual, "application/octet-stream")
		So(ff[0].Multiple(), ShouldEqual, true)
		So(ff[1].FieldName(), ShouldEqual, "file")
		So(ff[1].FileName(), ShouldEqual, "test2.txt")
		So(ff[1].ContentType(), ShouldEqual, "application/octet-stream")
		So(ff[1].Multiple(), ShouldEqual, true)
	})

	Convey("should not return any form file", t, func() {
		r := httptest.NewRequest("POST", "http://localhost", nil)
		req := request.New(request.Options{Request: r})
		ff := req.FormFiles()

		So(len(ff), ShouldEqual, 0)
	})

	Convey("should fail to return form files", t, func() {
		r := httptest.NewRequest("POST", "http://localhost", nil)
		r.Header.Set("Content-Type", "multipart/form-data")
		r.MultipartReader()
		req := request.New(request.Options{Request: r})
		ff := req.FormFiles()

		So(req.ContentType(), ShouldEqual, "multipart/form-data")
		So(len(ff), ShouldEqual, 0)
	})
}

func TestSaveFile(t *testing.T) {
	Convey("should save a form file", t, func() {
		f, err := os.Open("test.txt")
		defer f.Close()
		So(err, ShouldBeNil)

		buf := &bytes.Buffer{}
		mw := multipart.NewWriter(buf)
		w, err := mw.CreateFormFile("file", "test.txt")
		So(err, ShouldBeNil)
		_, err = io.Copy(w, f)
		So(err, ShouldBeNil)

		contentType := mw.FormDataContentType()
		mw.Close()

		r := httptest.NewRequest("POST", "http://localhost", buf)
		r.Header.Set("Content-Type", contentType)
		req := request.New(request.Options{Request: r})
		ff := req.FormFiles()

		tf := "./test.txt.tmp"
		defer os.Remove(tf)

		So(len(ff), ShouldEqual, 1)
		So(ff[0].SaveFile(tf, 644), ShouldBeNil)
	})

	Convey("should fail to save a form file (header)", t, func() {
		ff := request.FormFile{}
		So(ff.SaveFile("./test.txt.tmp", 644), ShouldBeError, errors.New("invalid file header"))
	})

	Convey("should fail to save a form file (open)", t, func() {
		f, err := os.Open("test.txt")
		defer f.Close()
		So(err, ShouldBeNil)

		buf := &bytes.Buffer{}
		mw := multipart.NewWriter(buf)
		w, err := mw.CreateFormFile("file", "test.txt")
		So(err, ShouldBeNil)
		_, err = io.Copy(w, f)
		So(err, ShouldBeNil)

		contentType := mw.FormDataContentType()
		mw.Close()

		r := httptest.NewRequest("POST", "http://localhost", buf)
		r.Header.Set("Content-Type", contentType)
		req := request.New(request.Options{Request: r, MaxMemory: -1})
		ff := req.FormFiles()

		tf := reflect.ValueOf(r.MultipartForm.File["file"][0]).Elem().FieldByName("tmpfile").String()
		if tf != "" {
			os.Remove(tf)
		}

		So(len(ff), ShouldEqual, 1)
		So(ff[0].SaveFile("./test.txt.tmp", 644), ShouldBeError, fmt.Errorf("failed to save file due to open %s: no such file or directory", tf))
	})

	Convey("should fail to save a form file (invalid file name)", t, func() {
		f, err := os.Open("test.txt")
		defer f.Close()
		So(err, ShouldBeNil)

		buf := &bytes.Buffer{}
		mw := multipart.NewWriter(buf)
		w, err := mw.CreateFormFile("file", "test.txt")
		So(err, ShouldBeNil)
		_, err = io.Copy(w, f)
		So(err, ShouldBeNil)

		contentType := mw.FormDataContentType()
		mw.Close()

		r := httptest.NewRequest("POST", "http://localhost", buf)
		r.Header.Set("Content-Type", contentType)
		req := request.New(request.Options{Request: r})
		ff := req.FormFiles()

		So(len(ff), ShouldEqual, 1)
		So(ff[0].SaveFile("", 644), ShouldBeError, errors.New("invalid file name"))
		So(ff[0].SaveFile("./test.txt/test.txt", 644), ShouldBeError, errors.New("failed to save file due to mkdir test.txt: not a directory"))
		defer os.Remove("./tmp/")
		So(ff[0].SaveFile("./tmp/.", 644), ShouldBeError, errors.New("failed to save file due to open tmp/.: is a directory"))
	})
}

func TestCopyTo(t *testing.T) {
	Convey("should copy to", t, func() {
		f, err := os.Open("test.txt")
		defer f.Close()
		So(err, ShouldBeNil)

		buf := &bytes.Buffer{}
		mw := multipart.NewWriter(buf)
		w, err := mw.CreateFormFile("file", "test.txt")
		So(err, ShouldBeNil)
		_, err = io.Copy(w, f)
		So(err, ShouldBeNil)

		contentType := mw.FormDataContentType()
		mw.Close()

		r := httptest.NewRequest("POST", "http://localhost", buf)
		r.Header.Set("Content-Type", contentType)
		req := request.New(request.Options{Request: r})
		ff := req.FormFiles()

		tf := "./test.txt.tmp"
		defer os.Remove(tf)

		So(len(ff), ShouldEqual, 1)
		n, err := ff[0].CopyTo(ioutil.Discard)
		So(n, ShouldBeGreaterThan, 0)
		So(err, ShouldBeNil)
	})

	Convey("should fail to copy to", t, func() {
		f, err := os.Open("test.txt")
		defer f.Close()
		So(err, ShouldBeNil)

		buf := &bytes.Buffer{}
		mw := multipart.NewWriter(buf)
		w, err := mw.CreateFormFile("file", "test.txt")
		So(err, ShouldBeNil)
		_, err = io.Copy(w, f)
		So(err, ShouldBeNil)

		contentType := mw.FormDataContentType()
		mw.Close()

		r := httptest.NewRequest("POST", "http://localhost", buf)
		r.Header.Set("Content-Type", contentType)
		req := request.New(request.Options{Request: r, MaxMemory: -1})
		ff := req.FormFiles()

		tf := reflect.ValueOf(r.MultipartForm.File["file"][0]).Elem().FieldByName("tmpfile").String()
		if tf != "" {
			os.Remove(tf)
		}

		So(len(ff), ShouldEqual, 1)
		n, err := ff[0].CopyTo(ioutil.Discard)
		So(n, ShouldEqual, 0)
		So(err, ShouldBeError, fmt.Errorf("failed to write due to open %s: no such file or directory", tf))
	})
}
