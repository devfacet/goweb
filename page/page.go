/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package page provides functions for handling web pages
package page

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

// Options represents the options than can be set when creating a new page
type Options struct {
	// URLPath holds the url path
	URLPath string
	// MatchAll matches everything after slash
	MatchAll bool
	// FilePath holds the file path
	FilePath string
	// FileSystem holds the file system
	FileSystem *http.FileSystem
	// Content holds the page content
	Content string
	// TemplateData holds the template data
	TemplateData interface{}
}

// New returns a page by the given options
func New(o Options) (*Page, error) {
	// Init the page
	page := Page{
		isInit:       true,
		urlPath:      o.URLPath,
		matchAll:     o.MatchAll,
		filePath:     o.FilePath,
		fileSystem:   o.FileSystem,
		content:      o.Content,
		templateData: o.TemplateData,
	}

	// Check vars
	if page.urlPath == "" {
		return nil, errors.New("invalid url path")
	}

	// If the file path is not empty then
	if page.filePath != "" {
		// Read the file and set the template content
		if page.fileSystem != nil {
			f, err := (*page.fileSystem).Open(page.filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to open file due to %s", err.Error())
			}
			defer f.Close()

			b, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, fmt.Errorf("failed to read file due to %s", err.Error())
			}
			page.content = string(b)
		} else {
			// TODO: Implement local file read
			return nil, fmt.Errorf("invalid file system")
		}
	}

	// If the content is not empty then
	if page.content != "" {
		var err error
		page.template, err = template.New(page.urlPath).Parse(page.content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template due to %s", err.Error())
		}
	}

	return &page, nil
}

// Page represents a web page
type Page struct {
	isInit       bool
	urlPath      string
	matchAll     bool
	filePath     string
	fileSystem   *http.FileSystem
	content      string
	templateData interface{}
	template     *template.Template
}

// URLPath returns the url path
func (page *Page) URLPath() string {
	return page.urlPath
}

// MatchAll returns whether the page url path should match all or not
func (page *Page) MatchAll() bool {
	return page.matchAll
}

// TemplateExecute executes the template by the given arguments
// TODO: add 2nd parameter for templateData and use page's templateData if it's nil
func (page *Page) TemplateExecute(w io.Writer, data interface{}) error {
	// If the template is nil then
	if page.template == nil {
		return nil
	}

	// If the given data is nil and the template data is not then
	if data == nil && page.templateData != nil {
		data = page.templateData // use the template data
	}

	// Execute the template
	if err := page.template.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template due to %s", err.Error())
	}

	return nil
}
