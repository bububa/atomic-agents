package document

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/atomic"
)

type Http struct {
	closed  *atomic.Bool
	fetched *atomic.Bool
	client  *http.Client
	httpReq *http.Request
	url     string
	reader  *bytes.Reader
	offset  int64
	size    int64
	mu      sync.Mutex
	Content
}

var (
	_ ParserReader = (*Http)(nil)
	_ fs.File      = (*Http)(nil)
)

type HttpConfig struct {
	client  *http.Client
	url     string
	method  string
	payload io.Reader
}

type HttpOption func(*HttpConfig)

func WithHttpMethod(method string) HttpOption {
	return func(h *HttpConfig) {
		h.method = method
	}
}

func WithHttpURL(url string) HttpOption {
	return func(h *HttpConfig) {
		h.url = url
	}
}

func WithPayload(payload io.Reader) HttpOption {
	return func(h *HttpConfig) {
		h.payload = payload
	}
}

func WithHttpClient(client *http.Client) HttpOption {
	return func(h *HttpConfig) {
		h.client = client
	}
}

func NewHttp(opts ...HttpOption) (*Http, error) {
	var cfg HttpConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.method == "" {
		cfg.method = http.MethodGet
	}
	if cfg.client == nil {
		cfg.client = http.DefaultClient
	}
	httpReq, err := http.NewRequest(cfg.method, cfg.url, cfg.payload)
	if err != nil {
		return nil, err
	}
	return &Http{
		closed:  atomic.NewBool(false),
		fetched: atomic.NewBool(false),
		client:  cfg.client,
		httpReq: httpReq,
		url:     cfg.url,
		reader:  new(bytes.Reader),
		Content: Content{
			meta: map[string]string{
				"url":    cfg.url,
				"method": cfg.method,
			},
		},
	}, nil
}

func (h *Http) fetch() error {
	if h.fetched.Swap(true) {
		return nil
	}
	resp, err := h.client.Do(h.httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	h.size = int64(len(bs))
	h.reader = bytes.NewReader(bs)
	return nil
}

// Read reads up to len(p) bytes into p.
func (h *Http) Read(p []byte) (n int, err error) {
	if h.closed.Load() {
		return 0, os.ErrClosed
	}
	if err := h.fetch(); err != nil {
		return 0, err
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.reader.Seek(h.offset, io.SeekStart)
	n, err = h.reader.Read(p)
	if err != nil {
		return 0, err
	}
	h.offset += int64(n)
	return
}

// Seek sets the offset for the next Read or Write on file to offset, interpreted
// according to whence: 0 means relative to the origin of the file, 1 means relative
// to the current offset, and 2 means relative to the end.
func (h *Http) Seek(offset int64, whence int) (int64, error) {
	if h.closed.Load() {
		return 0, os.ErrClosed
	}

	if err := h.fetch(); err != nil {
		return 0, err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = h.offset + offset
	case io.SeekEnd:
		// Seeking from the end is not supported in a streaming scenario easily
		return 0, fmt.Errorf("seeking from the end is not supported for lazy - loaded HTTP files")
	default:
		return 0, fmt.Errorf("invalid whence value: %d", whence)
	}

	if newOffset < 0 {
		return 0, fmt.Errorf("negative offset: %d", newOffset)
	}

	h.offset = newOffset
	return newOffset, nil
}

// ReadAt implements the io.ReaderAt interface.
func (h *Http) ReadAt(p []byte, off int64) (n int, err error) {
	if h.closed.Load() {
		return 0, os.ErrClosed
	}
	if err := h.fetch(); err != nil {
		return 0, err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if off >= int64(h.reader.Len()) {
		return 0, io.EOF
	}

	return h.reader.ReadAt(p, off)
}

func (h *Http) Close() error {
	h.closed.Store(true)
	return nil
}

// Stat returns the FileInfo structure describing file.
func (h *Http) Stat() (os.FileInfo, error) {
	if h.closed.Load() {
		return nil, os.ErrClosed
	}
	if err := h.fetch(); err != nil {
		return nil, err
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	return &FileInfo{
		name:    h.httpReq.URL.String(),
		size:    h.size,
		mode:    os.ModePerm,
		modTime: time.Now(),
	}, nil
}

func (h *Http) Size() int64 {
	stat, err := h.Stat()
	if err != nil {
		return 0
	}
	return stat.Size()
}
