package splitter

import (
	"bytes"
	"strings"
	"testing"
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
			input:     "basic chunking one.   chunking two? chunking three!.",
			chunkSize: 1,
			overlap:   0,
			wantChunks: []string{
				"basic chunking one.",
				"chunking two?",
				"chunking three!.",
			},
		},
		{
			name:       "basic chunking one 2",
			input:      "basic chunking one.   chunking two? chunking three!.",
			chunkSize:  4,
			overlap:    0,
			wantChunks: []string{"basic chunking one.", "chunking two? chunking three!."},
		},
		{
			name:       "with overlap",
			input:      "basic chunking one.   chunking two? chunking three!.",
			chunkSize:  4,
			overlap:    1,
			wantChunks: []string{"basic chunking one.", "basic chunking one. chunking two?", "chunking two? chunking three!."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
