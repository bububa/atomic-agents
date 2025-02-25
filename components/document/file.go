package document

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strconv"

	"go.uber.org/atomic"
)

type File struct {
	status *atomic.Int32
	fp     *os.File
	Document
}

var (
	_ ReadableDocument = (*File)(nil)
	_ ClosableDocument = (*File)(nil)
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
		status: atomic.NewInt32(Unread),
		fp:     fp,
		Document: Document{
			buffer: new(bytes.Buffer),
			meta: map[string]string{
				"filename": fileInfo.Name(),
				"modtime":  strconv.FormatInt(fileInfo.ModTime().Unix(), 10),
			},
		},
	}, nil
}

func (d *File) ReadStatus() ReadStatus {
	return d.status.Load()
}

func (d *File) ReadAll() error {
	if d.ReadStatus() == Reading {
		return ErrReading
	} else if d.ReadStatus() == ReadCompleted {
		return nil
	}
	if _, err := io.Copy(d.buffer, d.fp); err != nil {
		d.status.Store(Unread)
		return err
	}
	d.status.Store(ReadCompleted)
	return nil
}

func (d *File) Read() (chan<- []byte, error) {
	ch := make(chan<- []byte)
	if d.ReadStatus() == Reading {
		return nil, ErrReading
	} else if d.ReadStatus() == ReadCompleted {
		go func() {
			defer close(ch)
			d.status.Store(Reading)
			reader := bytes.NewReader(d.buffer.Bytes())
			tmp := make([]byte, 1024)
			for {
				n, err := reader.Read(tmp)
				if err != nil {
					d.status.Store(ReadCompleted)
					return
				}
				bs := make([]byte, n)
				copy(bs, tmp[:n])
				ch <- bs
			}
		}()
		return ch, nil
	}
	go func() {
		defer close(ch)
		d.status.Store(Reading)
		tmp := make([]byte, 1024)
		for {
			n, err := d.fp.Read(tmp)
			if err != nil {
				if errors.Is(err, io.EOF) {
					d.status.Store(ReadCompleted)
				} else {
					d.buffer.Reset()
					d.status.Store(Unread)
				}
				return
			}
			d.buffer.Write(tmp[:n])
		}
	}()
	return ch, nil
}

func (d *File) Close() error {
	return d.fp.Close()
}
