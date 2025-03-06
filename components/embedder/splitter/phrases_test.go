package splitter

import (
	"bytes"
	"strings"
	"testing"
)

func TestPhrases(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		chunkSize  int
		overlap    int
		wantChunks []string
		wantErr    bool
	}{
		{
			name:       "basic chunking",
			input:      "Hello, 世界. Nice — and totally adorable — dog; perhaps the “best one”! 🏆 🐶",
			chunkSize:  3,
			overlap:    0,
			wantChunks: []string{"Hello , 世", "界 . Nice", "— and totally", "adorable — dog", "; perhaps the", "“ best one", "” ! 🏆", "🐶"},
		},
		{
			name:       "with overlap",
			input:      "Hello, 世界. Nice — and totally adorable — dog; perhaps the “best one”! 🏆 🐶",
			chunkSize:  3,
			overlap:    1,
			wantChunks: []string{"Hello , 世", "世 界 .", ". Nice —", "— and totally", "totally adorable —", "— dog ;", "; perhaps the", "the “ best", "best one ”", "” ! 🏆", "🏆 🐶"},
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
