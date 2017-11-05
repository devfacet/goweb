/*
 * goweb
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

package request

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
)

// FormFile represents a form file
type FormFile struct {
	fieldName   string
	fileName    string
	contentType string
	multiple    bool
	fileHeader  *multipart.FileHeader
}

// FieldName returns the field name of the form file
func (ff *FormFile) FieldName() string {
	return ff.fieldName
}

// FileName returns the file name of the form file
func (ff *FormFile) FileName() string {
	return ff.fileName
}

// ContentType returns the content type of the form file
func (ff *FormFile) ContentType() string {
	return ff.contentType
}

// Multiple returns the multiple value of the form file
func (ff *FormFile) Multiple() bool {
	return ff.multiple
}

// FileHeader returns the file header for the form file
func (ff *FormFile) FileHeader() *multipart.FileHeader {
	return ff.fileHeader
}

// SaveFile saves the form file
func (ff *FormFile) SaveFile(name string) error {
	if name == "" {
		return errors.New("invalid file name")
	}

	if ff.fileHeader == nil {
		return errors.New("invalid file header")
	}

	// Open form file
	mf, err := ff.fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to save file due to %s", err.Error())
	}
	defer mf.Close()

	// Create file
	base := path.Base(name)
	dir := path.Dir(name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to save file due to %s", err.Error())
	}
	nf, err := os.Create(fmt.Sprintf("%s/%s", dir, base))
	if err != nil {
		return fmt.Errorf("failed to save file due to %s", err.Error())
	}
	defer nf.Close()

	// Copy data
	_, err = io.Copy(nf, mf)
	return err
}

// WriteTo writes the form file to the given destination
func (ff *FormFile) WriteTo(dst io.Writer) (written int64, err error) {
	// Open form file
	mf, err := ff.fileHeader.Open()
	if err != nil {
		return 0, fmt.Errorf("failed to write due to %s", err.Error())
	}
	defer mf.Close()

	// Copy data
	written, err = io.Copy(dst, mf)
	return written, err
}
