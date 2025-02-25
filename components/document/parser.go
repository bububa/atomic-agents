package document

import (
	"bytes"
	"context"
	"io"
	"strings"
	"unicode"
)

type Parser interface {
	Parse(context.Context, *bytes.Reader, io.Writer) error
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
