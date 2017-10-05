/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCheck(t *testing.T) {
	Convey("check the logger", t, func() {
		So(Logger, ShouldNotBeNil)
	})
}
