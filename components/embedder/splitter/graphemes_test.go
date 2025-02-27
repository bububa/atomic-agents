package splitter

import (
	"bytes"
	"strings"
	"testing"
)

func TestGraphemes(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		chunkSize    int
		overlap      int
		tokenCounter string
		wantChunks   []string
		wantErr      bool
	}{
		{
			name:         "basic chunking",
			input:        "hello world test",
			chunkSize:    5,
			overlap:      0,
			tokenCounter: "char",
			wantChunks:   []string{"hello", " worl", "d tes", "t"},
		},
		{
			name:         "with overlap",
			input:        "hello world",
			chunkSize:    6,
			overlap:      2,
			tokenCounter: "char",
			wantChunks:   []string{"hello ", "o worl", "rld"},
		},
		{
			name:         "basic chunking with default token counter",
			input:        "hello world test",
			chunkSize:    2,
			overlap:      0,
			tokenCounter: "field",
			wantChunks:   []string{"hello w", "orld t", "est"},
		},
		{
			name:         "with overlap with default token counter",
			input:        "hello world, axxxx bxxx cxxx",
			chunkSize:    3,
			overlap:      2,
			tokenCounter: "field",
			wantChunks:   []string{"hello world, a", ", ax", "axx", "xxx", "xxx bxxx c", "x cx", "cxx", "xxx"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var counter TokenCounter
			if tt.tokenCounter == "field" {
				counter = new(WordsTokenCounter)
			} else {
				counter = new(GraphemesTokenCounter)
			}
			splitter := NewGraphemes(
				WithChunkSize(tt.chunkSize),
				WithOverlap(tt.overlap),
				WithBuffer(bytes.NewBuffer([]byte(tt.input))),
				WithTokenCounter(counter),
			)
			if err := splitter.Scan(); err != nil {
				t.Error(err)
				return
			}
			t.Log(strings.Join(splitter.Chunks(), "\", \""))
			if len(tt.wantChunks) != splitter.Size() {
				t.Errorf("invalid chunks, token counter:%s, want %d, got %d", tt.tokenCounter, len(tt.wantChunks), splitter.Size())
				return
			}
			for i, want := range tt.wantChunks {
				got := make([]byte, 1024)
				n, _ := splitter.Read(got)
				if string(got[:n]) != want {
					t.Errorf("invalid chunk:%d, token counter:%s, want %s, got %s", i, tt.tokenCounter, want, string(got[:n]))
				}
			}
		})
	}
}
