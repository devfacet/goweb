/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package request

// Success represents a successful HTTP response
type Success struct {
	ID         string      `json:"id,omitempty"`
	StatusCode int         `json:"statusCode,omitempty"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Internal   interface{} `json:"-"`
}
