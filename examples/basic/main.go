// A basic app
package main

import (
	"net/http"

	"github.com/devfacet/goweb/log"
	"github.com/devfacet/goweb/page"
	"github.com/devfacet/goweb/server"
)

var (
	pageTitle = "basic"
	pageDesc  = "A basic app"
	fs        = http.FileSystem(http.Dir("./"))
)

func main() {
	// Init the server
	web := server.New(server.Options{
		ID:      "web",
		Address: "localhost:3000",
	})

	// Pages
	pages := []page.Options{
		page.Options{
			URLPath:    "/",
			FilePath:   "/templates/index.html",
			FileSystem: &fs,
			TemplateData: PageData{
				Title:       pageTitle,
				Description: pageDesc,
				Content:     "Hello world",
			},
		},
		page.Options{
			URLPath:      "foo",
			Content:      "{{.Body}}",
			TemplateData: struct{ Body string }{Body: "foo"},
		},
		page.Options{
			URLPath:      "bar/",
			Content:      "{{.Body}}",
			MatchAll:     true,
			TemplateData: struct{ Body string }{Body: "bar"},
		},
	}

	for _, v := range pages {
		p, err := page.New(v)
		if err != nil {
			log.Logger.Fatal(err)
		}
		if err := web.AddPage(p); err != nil {
			log.Logger.Fatal(err)
		}
	}

	// Listen
	if err := web.Listen(); err != nil {
		log.Logger.Fatal(err)
	}
}

// PageData represents a page data
type PageData struct {
	Title       string
	Description string
	Content     string
}
