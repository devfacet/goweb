/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package content_test

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/devfacet/goweb/content"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("should return a new content", t, func() {
		c := content.New(content.Options{})
		So(c.ContentType(), ShouldEqual, "")
		So(c.KnownType(), ShouldEqual, "")

		c = content.New(content.Options{MimeType: "text/plain; charset=utf-8"})
		So(c.ContentType(), ShouldEqual, "text/plain")
		So(c.KnownType(), ShouldEqual, "txt")

		b := bytes.NewBufferString("name,number\nfoo,1\nbar,2\n")
		r := bufio.NewReader(b)
		c = content.New(content.Options{Reader: r, MimeType: "text/csv"})
		So(c.ContentType(), ShouldEqual, "text/csv")
		So(c.KnownType(), ShouldEqual, "csv")

		b = bytes.NewBufferString("name\tnumber\nfoo\t1\nbar\t2\n")
		r = bufio.NewReader(b)
		c = content.New(content.Options{Reader: r, MimeType: "text/tab-separated-values"})
		So(c.ContentType(), ShouldEqual, "text/tab-separated-values")
		So(c.KnownType(), ShouldEqual, "tsv")
	})
}
