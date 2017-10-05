/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package server implements a web server
package server

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/devfacet/goweb/log"
	"github.com/devfacet/goweb/page"
	"github.com/devfacet/goweb/request"
	"github.com/devfacet/goweb/route"
)

// Options represents the options than can be set when creating a new server
type Options struct {
	// ID of the server
	ID string
	// Address of the server
	Address string
	// PathPrefix holds HTTP path prefix
	PathPrefix string
}

// New returns a new web server by the given options
func New(o Options) *Server {
	// Init the server
	server := Server{
		isInit:     true,
		id:         o.ID,
		address:    o.Address,
		pathPrefix: o.PathPrefix,
		mux:        http.NewServeMux(),
		ctx:        context.Background(),
	}

	if server.address == "" {
		// Use a random port number
		rand.Seed(time.Now().UTC().UnixNano())
		server.address = fmt.Sprintf("localhost:%d", rand.Intn(65535-3000)+3000)
	}

	if server.id == "" {
		// Use md5 checksum for server id
		server.id = fmt.Sprintf("%x", md5.Sum([]byte(server.address)))
	}

	if server.pathPrefix != "" {
		server.pathRoot = fmt.Sprintf("/%s/", strings.Trim(server.pathPrefix, "/"))
	}

	return &server
}

// ListenAll invokes listen for the given servers
func ListenAll(s ...*Server) error {
	// Init vars
	var result error
	sl := len(s)
	wg := sync.WaitGroup{}
	wg.Add(sl)

	// Iterate over the servers
	for _, v := range s {
		go func(v *Server) {
			// Listen
			if err := v.Listen(); err != nil && err != http.ErrServerClosed {
				// If there is no error before then
				if result == nil {
					result = err // set error
					// Cleanup all
					for i := 0; i < sl; i++ {
						wg.Done()
					}
				}
			} else {
				// Otherwise if there is no error then
				if result == nil {
					wg.Done() // done with the server
				}
			}
		}(v)
	}
	wg.Wait()

	// Shutdown all
	for i := 0; i < sl; i++ {
		s[i].Shutdown()
	}

	return result
}

// Server represents a web server
type Server struct {
	isInit     bool
	id         string
	address    string
	pathPrefix string
	pathRoot   string
	pages      []*page.Page
	http       *http.Server
	mux        *http.ServeMux
	ctx        context.Context
}

// ID returns the server id
func (server *Server) ID() string {
	return server.id
}

// Address returns the server address
func (server *Server) Address() string {
	return server.address
}

// PathRoot returns the server path prefix
func (server *Server) PathRoot() string {
	return server.pathRoot
}

// Listen initializes the server and listens for requests
func (server *Server) Listen() error {
	// Route list
	for _, v := range server.Routes() {
		log.Logger.Printf("route definition: %s > %s - explicit:%t, redirect:%t", v.Path(), v.Pattern(), v.Explicit(), v.Redirect())
	}

	// Listen
	server.http = &http.Server{Addr: server.address, Handler: server.mux, ErrorLog: log.Logger}
	var err error
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		log.Logger.Printf("%s listening on %s", server.id, server.address)
		err = server.http.ListenAndServe()
		wg.Done()
	}()
	wg.Wait()
	return err
}

// Close closes all active listeners and connections immediately
func (server *Server) Close() error {
	if server.http != nil {
		return server.http.Close()
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (server *Server) Shutdown() error {
	if server.http != nil {
		return server.http.Shutdown(server.ctx)
	}
	return nil
}

// Routes returns the list of the routes
func (server *Server) Routes() []route.Route {
	return route.ListByMux(server.mux)
}

// AddHandler adds a handler
func (server *Server) AddHandler(pattern string, handler http.Handler) {
	if server.pathRoot != "" {
		pattern = fmt.Sprintf("%s%s", server.pathRoot, strings.TrimLeft(pattern, "/"))
	} else {
		pattern = fmt.Sprintf("/%s", strings.TrimLeft(pattern, "/"))
	}
	server.mux.Handle(pattern, http.StripPrefix(pattern, handler))
}

// AddHandlerFunc adds a handler function
func (server *Server) AddHandlerFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if server.pathRoot != "" {
		pattern = fmt.Sprintf("%s%s", server.pathRoot, strings.TrimLeft(pattern, "/"))
	} else {
		pattern = fmt.Sprintf("/%s", strings.TrimLeft(pattern, "/"))
	}
	server.mux.HandleFunc(pattern, handler)
}

// AddPage adds a page
func (server *Server) AddPage(p *page.Page) error {
	// Init vars
	puf := p.URLPath()

	if puf == "" {
		return errors.New("invalid page url")
	}

	if server.pathRoot != "" {
		puf = fmt.Sprintf("%s%s", server.pathRoot, strings.TrimLeft(puf, "/"))
	}
	ts := strings.HasSuffix(puf, "/")

	// Define handler function
	h := func(w http.ResponseWriter, r *http.Request) {
		// If match all is false, page url has trailing slash and requested url path is not equal to page url then
		if !p.MatchAll() && ts && puf != r.URL.Path {
			request.New(request.Options{Request: r, Writer: w}).Reply(request.Error{
				StatusCode: 404,
			})
			return
		}

		// Execute the template and write into response
		if err := p.TemplateExecute(w, nil); err != nil {
			request.New(request.Options{Request: r, Writer: w}).Reply(request.Error{
				StatusCode: 500,
				Internal:   err,
			})
			return
		}
	}
	server.AddHandlerFunc(p.URLPath(), h)

	return nil
}

// AddPages adds pages
func (server *Server) AddPages(p ...*page.Page) error {
	// Iterate over pages
	for _, v := range p {
		if err := server.AddPage(v); err != nil {
			return err
		}
	}

	return nil
}
