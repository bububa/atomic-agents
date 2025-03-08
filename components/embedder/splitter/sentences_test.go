package splitter

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/clipperhouse/uax29/sentences"
)

func TestSentences(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		chunkSize  int
		overlap    int
		wantChunks []string
		wantErr    bool
	}{
		{
			name:      "basic chunking one",
			input:     "Basic chunking one. Chunking two? Chunking three!",
			chunkSize: 1,
			overlap:   0,
			wantChunks: []string{
				"Basic chunking one.",
				"Chunking two?",
				"Chunking three!",
			},
		},
		{
			name:       "basic chunking one 2",
			input:      "Basic chunking one. Chunking two? Chunking three!",
			chunkSize:  9,
			overlap:    0,
			wantChunks: []string{"Basic chunking one. Chunking two?", "Chunking three!"},
		},
		{
			name:       "with overlap",
			input:      "Basic chunking one. Chunking two? Chunking three!",
			chunkSize:  4,
			overlap:    1,
			wantChunks: []string{"Basic chunking one.", "Basic chunking one. Chunking two?", "Chunking two? Chunking three!"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segmenter := sentences.NewSegmenter([]byte(tt.input))
			for segmenter.Next() {
				fmt.Println(segmenter.Text())
			}
			splitter := NewSentences(
				WithChunkSize(tt.chunkSize),
				WithOverlap(tt.overlap),
				WithBuffer(bytes.NewBuffer([]byte(tt.input))),
				WithTokenCounter(new(WordsTokenCounter)),
			)
			if err := splitter.Scan(); err != nil {
				t.Error(err)
				return
			}
			t.Log(strings.Join(splitter.Chunks(), "\", \""))
			if len(tt.wantChunks) != splitter.Size() {
				t.Errorf("invalid chunks, want %d, got %d", len(tt.wantChunks), splitter.Size())
				return
			}
			for i, want := range tt.wantChunks {
				got := make([]byte, 1024)
				n, _ := splitter.Read(got)
				if string(got[:n]) != want {
					t.Errorf("invalid chunk:%d, want %s, got %s", i, want, string(got[:n]))
				}
			}
		})
	}
}
