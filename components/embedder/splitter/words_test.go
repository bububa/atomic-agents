package splitter

import (
	"bytes"
	"strings"
	"testing"
)

func TestWords(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		chunkSize  int
		overlap    int
		wantChunks []string
		wantErr    bool
	}{
		{
			name:      "basic chunking",
			input:     "one two three four five six seven eight nine ten.",
			chunkSize: 3,
			overlap:   0,
			wantChunks: []string{
				"one two three",
				"four five six",
				"seven eight nine",
				"ten.",
			},
		},
		{
			name:      "with overlap",
			input:     "one two three four five six seven eight nine ten.",
			chunkSize: 3,
			overlap:   1,
			wantChunks: []string{
				"one two three",
				"three four five",
				"five six seven",
				"seven eight nine",
				"nine ten.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitter := NewWords(
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
