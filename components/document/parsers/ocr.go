package parsers

import "io"

type OCR interface {
	Run(r io.Reader) (string, error)
	Close() error
}
