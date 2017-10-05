/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package request provides functions for handling HTTP requests
package request

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type contextKey string

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

var (
	// Logger holds the global logger that can be override by another logger
	Logger *log.Logger

	// ContextKeys holds request context keys
	ContextKeys = struct {
		PathPrefix contextKey
	}{
		PathPrefix: "PathPrefix",
	}
)

func init() {
	// Init logger
	Logger = log.New(os.Stdout, "", log.LstdFlags)
	// If it's a test then
	if flag.Lookup("test.v") != nil {
		Logger.SetOutput(ioutil.Discard) // discard logs
	}
}

// Options represents the options than can be set when creating a new request
type Options struct {
	// Request holds the request
	Request *http.Request
	// Writer holds the response writer
	Writer http.ResponseWriter
	// MaxMemory holds the maximum memory for multi part form parsing
	MaxMemory int64
}

// New returns a new HTTP request by the given options
func New(o Options) *Request {
	// Init the request
	request := Request{
		isInit:    true,
		r:         o.Request,
		w:         o.Writer,
		maxMemory: o.MaxMemory,
	}

	if request.r == nil {
		request.r = &http.Request{}
	}

	if request.w == nil {
		request.w = *new(http.ResponseWriter)
	}

	if request.maxMemory == 0 {
		request.maxMemory = defaultMaxMemory
	}

	// Check content type
	if request.ContentType() == "application/json" || request.ContentType() == "application/javascript" {
		request.isJSON = true
	}

	return &request
}

// Request represents an HTTP request
type Request struct {
	isInit      bool
	w           http.ResponseWriter
	r           *http.Request
	maxMemory   int64
	contentType string
	isJSON      bool
	isError     bool
	result      []byte
}

// ContentType returns the request content type
func (request *Request) ContentType() string {
	if request.contentType == "" && request.r != nil {
		request.contentType = request.r.Header.Get("Content-Type")
		if request.contentType != "" {
			request.contentType = strings.SplitN(request.contentType, ";", 2)[0]
		}
	}
	return request.contentType
}

// Reply replies an HTTP request
func (request *Request) Reply(rv interface{}) {
	// Init vars
	var jsonData interface{}
	jsonTrig := false
	result := []byte{}
	header := http.StatusOK

	// Determine type
	if rv != nil {
		switch rv.(type) {
		case []uint8:
			result = rv.([]byte)
		case string:
			result = []byte(rv.(string))
		case int:
			result = []byte(strconv.Itoa(rv.(int)))
		case int64:
			result = []byte(strconv.FormatInt(rv.(int64), 10))
		case float64:
			result = []byte(strconv.FormatFloat(rv.(float64), 'f', -1, 64))
		case Success:
			s := rv.(Success)
			header = http.StatusOK // default

			// Check and update status
			if s.StatusCode != 0 {
				if st := http.StatusText(s.StatusCode); st != "" {
					// Set header
					header = s.StatusCode
					// Check and update message
					if s.Message == "" {
						s.Message = st
					}
				}
			} else {
				// Default success
				s.StatusCode = http.StatusOK
				s.Message = http.StatusText(s.StatusCode)
			}

			jsonTrig = true
			jsonData = s
		case Error:
			e := rv.(Error)
			header = http.StatusInternalServerError // default

			// Check and update status
			if e.StatusCode != 0 {
				if st := http.StatusText(e.StatusCode); st != "" {
					// Set header
					header = e.StatusCode
					// Check and update message
					if e.Message == "" {
						e.Message = st
					}
				}
			} else {
				// Default error
				e.StatusCode = http.StatusInternalServerError
				e.Message = http.StatusText(e.StatusCode)
			}

			jsonTrig = true
			jsonData = e
		default:
			t := reflect.Indirect(reflect.ValueOf(rv)).Kind().String()

			// If it's a struct then
			if t == "struct" {
				jsonTrig = true
				jsonData = rv
			} else {
				// Otherwise
				Logger.Printf("unknown type: %T/%s/%s", rv, reflect.ValueOf(rv).Kind().String(), reflect.Indirect(reflect.ValueOf(rv)).Kind().String())
				result = []byte(fmt.Sprintf("%s", rv))
			}
		}
	}

	if jsonTrig {
		// Set content type
		cb := request.r.URL.Query().Get("callback")
		if cb != "" {
			request.w.Header().Set("Content-Type", "application/javascript") // JSONP
		} else {
			request.w.Header().Set("Content-Type", "application/json")
		}

		// Prepare the result
		var err error
		if result, err = json.Marshal(jsonData); err != nil {
			Logger.Printf("failed to reply due to parse error: %s", err.Error())
			header = http.StatusInternalServerError
			result = []byte(fmt.Sprintf(`{"statusCode":%d,"message":"%s"}`, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)))
		}

		if cb != "" {
			result = []byte(fmt.Sprintf("%s(%s)", cb, result))
		}
		request.w.WriteHeader(header)
	}

	request.w.Write(result)
}

// FormFiles returns the form files
func (request *Request) FormFiles() []FormFile {
	// Init vars
	result := []FormFile{}

	if request.ContentType() != "multipart/form-data" {
		return result
	}

	if err := request.r.ParseMultipartForm(request.maxMemory); err != nil {
		Logger.Printf("failed to parse multipart form due to %s", err.Error())
		return result
	}

	// Iterate over files and add to result
	for k, v := range request.r.MultipartForm.File {
		// Iterate over values
		for _, vv := range v {
			ff := FormFile{
				fieldName:   k,
				fileName:    vv.Filename,
				contentType: vv.Header.Get("Content-Type"),
				fileHeader:  vv,
			}
			if len(v) > 1 {
				ff.multiple = true
			}
			result = append(result, ff)
		}
	}

	return result
}
