package splitter

import (
	"bufio"
	"bytes"
	"io"

	"github.com/bububa/atomic-agents/components/embedder"
)

type Scanner interface {
	Bytes() []byte
	Text() string
	Scan() bool
	Err() error
}

type Options struct {
	chunkSize    int
	overlap      int
	rw           *bytes.Buffer
	scanner      Scanner
	tokenCounter TokenCounter
	delimiter    []byte
	chunks       [][]byte
	offset       int
}

var _ embedder.Chunker = (*Options)(nil)

// Option is a function type for configuring chunkcer Options.
// This follows the functional options pattern for clean and flexible configuration.
type Option func(*Options)

func WithChunkSize(size int) Option {
	return func(o *Options) {
		o.chunkSize = size
	}
}

func WithOverlap(overlap int) Option {
	return func(o *Options) {
		o.overlap = overlap
	}
}

func WithBuffer(rw *bytes.Buffer) Option {
	return func(o *Options) {
		o.rw = rw
	}
}

func WithTokenCounter(counter TokenCounter) Option {
	return func(o *Options) {
		o.tokenCounter = counter
	}
}

func (o *Options) Buffer() *bytes.Buffer {
	if o.rw == nil {
		o.rw = new(bytes.Buffer)
	}
	return o.rw
}

func (o *Options) Scanner() Scanner {
	if o.scanner == nil {
		o.scanner = bufio.NewScanner(o.Buffer())
	}
	return o.scanner
}

func (o *Options) Write(p []byte) (int, error) {
	n := len(p)
	dist := make([]byte, n)
	n = copy(dist, p)
	o.chunks = append(o.chunks, dist)
	return n, nil
}

func (o *Options) Read(p []byte) (int, error) {
	l := len(o.chunks)
	if o.offset >= l {
		return 0, io.EOF
	}
	n := copy(p, o.chunks[o.offset])
	o.offset++
	return n, nil
}

func (o *Options) Chunks() []string {
	ret := make([]string, len(o.chunks))
	for idx, v := range o.chunks {
		ret[idx] = string(v)
	}
	return ret
}

func (o *Options) Size() int {
	return len(o.chunks)
}

func (o *Options) Scan() error {
	var parts [][]byte
	var currentChunk Chunk
	var currentTokenCount int
	for i := 0; o.scanner.Scan(); i++ {
		bs := o.scanner.Bytes()
		part := make([]byte, len(bs))
		copy(part, bs)
		parts = append(parts, part)
		partTokenCount := o.tokenCounter.Count(part)
		if currentTokenCount+partTokenCount > o.chunkSize && currentTokenCount > 0 {
			overlapStart := max(currentChunk.Start, currentChunk.End-o.estimateOverlapParts(parts, currentChunk.End, o.overlap))
			if _, err := currentChunk.Buffer.WriteTo(o); err != nil {
				return err
			}
			o.joinPartsWithoutDelimiter(currentChunk.Buffer, parts[overlapStart:i+1])
			currentChunk.TokenSize = 0
			currentChunk.Start = overlapStart
			currentChunk.End = i + 1
			currentTokenCount = 0
			for j := overlapStart; j <= i; j++ {
				currentTokenCount += o.tokenCounter.Count(parts[j])
			}
		} else {
			if currentTokenCount == 0 {
				currentChunk.Start = i
			}
			if !bytes.Equal(part, o.delimiter) {
				if i > 0 {
					currentChunk.Buffer.Write(o.delimiter)
				}
				part = bytes.TrimSuffix(part, o.delimiter)
				if currentChunk.Buffer == nil {
					currentChunk.Buffer = bytes.NewBuffer(part)
				} else {
					currentChunk.Buffer.Write(part)
				}
			}
			currentChunk.End = i + 1
			currentTokenCount = o.tokenCounter.Count(currentChunk.Buffer.Bytes())
		}
		currentChunk.TokenSize = currentTokenCount
	}
	if currentChunk.TokenSize > 0 {
		if _, err := currentChunk.Buffer.WriteTo(o); err != nil {
			return err
		}
	}

	return o.scanner.Err()
}

func (o *Options) TokenCount(txt string) int {
	return o.tokenCounter.Count([]byte(txt))
}

func (o *Options) SplitText(txt string) []string {
	o.rw.Reset()
	o.rw.WriteString(txt)
	if err := o.Scan(); err != nil {
		return nil
	}
	return o.Chunks()
}

func (o *Options) joinPartsWithoutDelimiter(w io.Writer, parts [][]byte) {
	for idx, part := range parts {
		part = bytes.Trim(part, string(o.delimiter))
		if len(part) == 0 {
			continue
		}
		if idx > 0 {
			w.Write(o.delimiter)
		}
		w.Write(part)
	}
}

// estimateOverlapParts calculates how many parts from the end of the
// previous chunk should be included in the next chunk to achieve the desired
// token overlap.
func (o *Options) estimateOverlapParts(parts [][]byte, endPart, desiredOverlap int) int {
	overlapTokens := 0
	overlapParts := 0
	for i := endPart - 1; i >= 0 && overlapTokens < desiredOverlap; i-- {
		overlapTokens += o.tokenCounter.Count(parts[i])
		overlapParts++
	}
	return overlapParts
}
