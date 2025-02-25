package pptx

import (
	"archive/zip"
	"bytes"
	"errors"
	"path"
	"regexp"
	"strings"
	"unsafe"

	qxml "github.com/dgrr/quickxml"
)

// ParseRelsMap parses a zip file and returns a mapping of relationship IDs to target strings.
//
// Parameters:
//   - f: *zip.File object representing the zip file to parse.
//   - prefix: string prefix used to construct full part name(target string).
//
// Returns:
//   - map[string]string: a mapping of relationship IDs to target strings.
//   - error: an error object indicating any error occurred during parsing.
func ParseRelsMap(f *zip.File, preffix string) (map[string]string, error) {
	m := make(map[string]string)
	if f == nil {
		return nil, errors.New("nil zip file")
	}
	rc, err := f.Open()
	if err != nil {
		return m, err
	}
	defer rc.Close()

	target := new(strings.Builder)

	r := qxml.NewReader(rc)

	for r.Next() {
		switch e := r.Element().(type) {
		case *qxml.StartElement:
			switch e.Name() {
			case "Relationship":
				attrs := e.Attrs()
				if attrs.Len() > 0 {
					rIdAttr := attrs.Get("Id")
					targetAttr := attrs.Get("Target")
					t := formatTarget(targetAttr.Value(), preffix)
					target.WriteString(t)
					m[rIdAttr.Value()] = target.String()
					target.Reset()
				}
			}
		}
	}

	return m, nil
}

// formatTarget formats the target string by adding the prefix if it doesn't have it already.
//
// Parameters:
//   - target: the string to be formatted.
//   - prefix: the prefix to be added to the target string.
//
// Returns:
//   - string: the formatted target string.
func formatTarget(target, preffix string) string {
	if strings.HasPrefix(target, preffix) {
		return target
	}
	t := path.Clean(target)
	t = strings.TrimPrefix(t, "../")

	return preffix + t
}

// MaxLineLenWithPrefix calculates the maximum line length in a string with a given prefix.
//
// Parameters:
//   - s: the input string
//   - prefix: the prefix to add to each line
//
// Returns:
//   - string: the modified string with the added prefix
//   - int: the maximum line length (including the prefix)
func MaxLineLenWithPrefix(s string, prefix []byte) (string, int) {
	maxLen := 0
	lineLen := 0
	newS := make([]byte, 0, len(s)+10)
	buf := bytes.NewBuffer(newS)
	for i, b := range StringTobytes(s) {
		if i == 0 {
			buf.Write(prefix)
		}

		if b != '\n' {
			lineLen++
			buf.WriteByte(b)
		} else {
			if lineLen > maxLen {
				maxLen = lineLen
			}
			lineLen = 0
			buf.WriteByte(b)
			buf.Write(prefix)
		}
	}

	if lineLen > maxLen {
		maxLen = lineLen
	}

	maxLen += len(prefix)

	return buf.String(), maxLen
}

// StringTobytes converts a string to a byte slice.
//
// It takes a string parameter `s` and returns a byte slice.
//
// This function is implemented using the `unsafe` package to achieve zero cost conversion.
func StringTobytes(s string) []byte {
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	return b
}

// MatchNameIterTo is a function that matches the name pattern and the to pattern
// iteratively using the given qxml.Reader. It returns true if the name pattern
// is matched and false if the to pattern is matched or if the end of the reader
// is reached.
//
// Parameters:
//   - r: A pointer to a qxml.Reader object
//   - namePattern: The regular expression pattern to match the name
//   - toPattern: The regular expression pattern to match the to
//
// Return:
//   - bool: true if the name pattern is matched, false otherwise
func MatchNameIterTo(r *qxml.Reader, namePattern string, toPattern string) bool {
	re_NAME := regexp.MustCompile(namePattern)
	re_TO := regexp.MustCompile(toPattern)

	for r.Next() {
		switch e := r.Element().(type) {
		case *qxml.StartElement:
			if re_NAME.MatchString(e.Name()) {
				return true
			}
		case *qxml.EndElement:
			if re_TO.MatchString(e.Name()) {
				return false
			}
		}
	}

	return false
}

// FindNameIterTo finds the given name iteratively in the qxml Reader until it reaches the specified end element.
//
// Parameters:
//   - r: a pointer to the qxml Reader.
//   - name: the name to search for in the qxml Reader.
//   - to: the end element name to stop the search.
//
// Returns:
//   - true if the name is found before reaching the end element, false otherwise.
func FindNameIterTo(r *qxml.Reader, name string, to string) bool {
	for r.Next() {
		switch e := r.Element().(type) {
		case *qxml.StartElement:
			if e.Name() == name {
				return true
			}
		case *qxml.EndElement:
			if e.Name() == to {
				return false
			}
		}
	}

	return false
}
