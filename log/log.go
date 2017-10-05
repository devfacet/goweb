/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package log implements a simple logging package
// Currently it's pretty much a place holder for future implementations.
package log

import (
	"flag"
	"io/ioutil"
	"log"
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
