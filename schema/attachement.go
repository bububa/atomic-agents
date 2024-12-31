package schema

import "io"

// Attachement message attachement
type Attachement struct {
	// ImageURL attached image_url
	ImageURLs []string `json:"image_url,omitempty"`
	// Files attached file
	Files []io.Reader `json:"file,omitempty"`
}
