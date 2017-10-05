/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package content provides functions for content handling
// such as content type detection.
package content

import (
	"bufio"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	// Logger holds the global logger that can be override by another logger
	Logger *log.Logger
)

func init() {
	// Init logger
	Logger = log.New(os.Stdout, "", log.LstdFlags)
	// If it's a test then
	if flag.Lookup("test.v") != nil {
		Logger.SetOutput(ioutil.Discard) // discard logs
	}
}

// Options represents the options than can be set when creating a new content
type Options struct {
	// Reader holds the reader
	Reader *bufio.Reader
	// MimeType holds the mime type
	MimeType string
}

// New returns a content by the given options
func New(o Options) *Content {
	// Init the content
	content := Content{
		isInit:   true,
		reader:   o.Reader,
		mimeType: o.MimeType,
	}
	content.detectType()

	return &content
}

// Content represents a content
type Content struct {
	isInit          bool
	reader          *bufio.Reader
	mimeType        string
	contentType     string
	contentTypeOrig string
	knownType       string
}

// ContentType returns the content type
func (c *Content) ContentType() string {
	return c.contentType
}

// KnownType returns the known type
func (c *Content) KnownType() string {
	return c.knownType
}

func (c *Content) detectType() {
	// REF: https://golang.org/src/net/http/sniff.go

	// Determine the content type by reading data from the reader
	data := make([]byte, 1024)
	if c.reader != nil {
		b, err := c.reader.Peek(1024) // peek 1024 bytes
		if err == nil || err == io.EOF {
			copy(data, b) // copy data for optimization
			c.contentTypeOrig = http.DetectContentType(data)
		}
	}

	// If the given mime type is not empty then
	if c.mimeType != "" {
		if c.contentTypeOrig == "application/octet-stream" || c.contentTypeOrig == "" {
			c.contentTypeOrig = c.mimeType
		}
	}
	c.contentType = strings.Split(c.contentTypeOrig, ";")[0] // example: text/plain; charset=utf-8

	// Text
	t := map[string]bool{
		"text":                      true,
		"text/plain":                true,
		"text/csv":                  true,
		"text/tab-separated-values": true,
	}
	if _, ok := t[c.contentType]; ok {
		l := strings.Split(string(data), "\n")
		ll := len(l)

		// Default known type
		c.knownType = "txt"

		// CSV
		// Check for at least two lines and match commas
		if ll > 1 && strings.Count(l[0], ",") > 0 && strings.Count(l[0], ",") == strings.Count(l[1], ",") {
			c.knownType = "csv"
		}

		// TSV
		// Check for at least two lines and match tabs
		if ll > 1 && strings.Count(l[0], "\t") > 0 && strings.Count(l[0], "\t") == strings.Count(l[1], "\t") {
			c.knownType = "tsv"
		}
	}
}
