package document

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"go.uber.org/atomic"
)

type Http struct {
	status  *atomic.Int32
	client  *http.Client
	httpReq *http.Request
	Document
}

var (
	_ ReadableDocument = (*File)(nil)
	_ ClosableDocument = (*File)(nil)
)

type HttpConfig struct {
	client  *http.Client
	link    string
	method  string
	payload io.Reader
}

type HttpOption func(*HttpConfig)

func WithHttpMethod(method string) HttpOption {
	return func(h *HttpConfig) {
		h.method = method
	}
}

func WithHttpURL(link string) HttpOption {
	return func(h *HttpConfig) {
		h.link = link
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
	httpReq, err := http.NewRequest(cfg.method, cfg.link, cfg.payload)
	if err != nil {
		return nil, err
	}
	return &Http{
		status:  atomic.NewInt32(Unread),
		client:  cfg.client,
		httpReq: httpReq,
		Document: Document{
			buffer: new(bytes.Buffer),
			meta: map[string]string{
				"url":    cfg.link,
				"method": cfg.method,
			},
		},
	}, nil
}

func (h *Http) ReadStatus() ReadStatus {
	return h.status.Load()
}

func (h *Http) ReadAll() error {
	if h.ReadStatus() == Reading {
		return ErrReading
	} else if h.ReadStatus() == ReadCompleted {
		return nil
	}
	httpResp, err := h.client.Do(h.httpReq)
	if err != nil {
		h.status.Store(Unread)
		return err
	}
	defer httpResp.Body.Close()
	if _, err = io.Copy(h.buffer, httpResp.Body); err != nil {
		h.status.Store(Unread)
	}
	h.status.Store(ReadCompleted)
	return nil
}

func (h *Http) Read() (chan<- []byte, error) {
	ch := make(chan<- []byte)
	if h.ReadStatus() == Reading {
		return nil, ErrReading
	} else if h.ReadStatus() == ReadCompleted {
		go func() {
			defer close(ch)
			h.status.Store(Reading)
			reader := bytes.NewReader(h.buffer.Bytes())
			tmp := make([]byte, 1024)
			for {
				n, err := reader.Read(tmp)
				if err != nil {
					h.status.Store(ReadCompleted)
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
		h.status.Store(Reading)
		httpResp, err := h.client.Do(h.httpReq)
		if err != nil {
			h.status.Store(Unread)
			return
		}
		defer httpResp.Body.Close()
		tmp := make([]byte, 1024)
		for {
			n, err := httpResp.Body.Read(tmp)
			if err != nil {
				if errors.Is(err, io.EOF) {
					h.status.Store(ReadCompleted)
				} else {
					h.buffer.Reset()
					h.status.Store(Unread)
				}
				return
			}
			h.buffer.Write(tmp[:n])
		}
	}()
	return ch, nil
}
