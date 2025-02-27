package embedder

type Chunker interface {
	SplitText(string) []string
	TokenCount(txt string) int
}
