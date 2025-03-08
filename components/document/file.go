package document

import (
	"errors"
	"io/fs"
	"os"
	"strconv"
)

type File struct {
	fp *os.File
	Content
}

var (
	_ ParserReader = (*File)(nil)
	_ fs.File      = (*File)(nil)
)

func NewFile(fname string) (*File, error) {
	fp, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	fileInfo, err := fp.Stat()
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		return nil, errors.New("FileDocument could not be a directory")
	}
	return &File{
		fp: fp,
		Content: Content{
			meta: map[string]string{
				"filename": fileInfo.Name(),
				"modtime":  strconv.FormatInt(fileInfo.ModTime().Unix(), 10),
			},
		},
	}, nil
}

func (d *File) Stat() (os.FileInfo, error) {
	return d.fp.Stat()
}

func (d *File) Read(p []byte) (int, error) {
	return d.fp.Read(p)
}

func (d *File) ReadAt(p []byte, off int64) (int, error) {
	return d.fp.ReadAt(p, off)
}

func (d *File) Size() int64 {
	stat, err := d.Stat()
	if err != nil {
		return 0
	}
	return stat.Size()
}

func (d *File) Close() error {
	return d.fp.Close()
}
