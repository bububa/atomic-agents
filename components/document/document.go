package document

import (
	"io"
	"io/fs"
	"time"
)

type Document interface {
	io.ReaderFrom
	io.Writer
	io.WriterAt
	io.WriterTo
	String() string
	Meta() map[string]string
}

// Content is a document container with metadata
type Content struct {
	content []byte
	meta    map[string]string
}

func (d *Content) String() string {
	return string(d.content)
}

func (d *Content) Bytes() []byte {
	return d.content
}

func (d *Content) Meta() map[string]string {
	return d.meta
}

func (d *Content) ReadFrom(r io.Reader) (int64, error) {
	// Create a buffer to read the data
	buffer := make([]byte, 1024)
	var n int64
	for {
		// Read data from the reader into the buffer
		numRead, err := r.Read(buffer)
		if numRead > 0 {
			// Append the read data to the content field
			d.content = append(d.content, buffer[:numRead]...)
			n += int64(numRead)
		}
		if err != nil {
			if err == io.EOF {
				// End of file reached, break the loop
				break
			}
			// Return the error if it's not EOF
			return n, err
		}
	}
	return n, nil
}

// Write implements the io.Writer interface.
// It appends the given data to the content of the document.
func (d *Content) Write(p []byte) (n int, err error) {
	d.content = append(d.content, p...)
	return len(p), nil
}

// WriteAt implements the io.WriterAt interface.
// It writes the given data at the specified offset in the document's content.
// If the offset is beyond the current content length, it pads with zeros.
func (d *Content) WriteAt(p []byte, off int64) (n int, err error) {
	contentLen := int64(len(d.content))
	if off > contentLen {
		padding := make([]byte, off-contentLen)
		d.content = append(d.content, padding...)
	}
	if int64(len(p))+off > contentLen {
		newLen := int64(len(p)) + off
		for int64(len(d.content)) < newLen {
			d.content = append(d.content, 0)
		}
	}
	copy(d.content[off:], p)
	return len(p), nil
}

// WriteTo implements the io.WriterTo interface.
// It writes the content of the document to the provided io.Writer.
func (d *Content) WriteTo(w io.Writer) (n int64, err error) {
	written, err := w.Write(d.content)
	return int64(written), err
}

// FileInfo represents the file information for an S3 object.
type FileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
}

// Name returns the base name of the file.
func (s *FileInfo) Name() string {
	return s.name
}

// Size returns the length in bytes for regular files; system - dependent for others.
func (s *FileInfo) Size() int64 {
	return s.size
}

// Mode returns the file mode bits.
func (s *FileInfo) Mode() fs.FileMode {
	return s.mode
}

// ModTime returns the modification time.
func (s *FileInfo) ModTime() time.Time {
	return s.modTime
}

// IsDir returns true if the file is a directory.
func (s *FileInfo) IsDir() bool {
	return false
}

// Sys returns the underlying data source (can return nil).
func (s *FileInfo) Sys() interface{} {
	return nil
}
