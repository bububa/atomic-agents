package document

import (
	"context"
	"io"
	"strings"
	"unicode"
)

type Parser interface {
	Parse(context.Context, ParserReader, io.Writer) error
}

type ParserReader interface {
	io.Reader
	io.ReaderAt
	Size() int64
}

// EscapeMarkdown escapes special characters in a string for Markdown
func EscapeMarkdown(s string) string {
	specialChars := []string{"\\", "`", "*", "_", "{", "}", "[", "]", "(", ")", "#", "+", "-", ".", "!", "~", "|"}
	for _, char := range specialChars {
		s = strings.ReplaceAll(s, char, "\\"+char)
	}
	return s
}

func StripUnprintable(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsPrint(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}
