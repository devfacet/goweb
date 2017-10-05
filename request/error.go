/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package request

// Error represents an HTTP error
type Error struct {
	ID         string      `json:"id,omitempty"`
	StatusCode int         `json:"statusCode,omitempty"`
	Message    string      `json:"message,omitempty"`
	Error      interface{} `json:"error,omitempty"`
	Internal   interface{} `json:"-"`
}
