package schema

import "io"

// Attachement message attachement
type Attachement struct {
	// ImageURL attached image_url
	ImageURLs []string `json:"image_url,omitempty"`
	// Files attached file
	Files []io.Reader `json:"file,omitempty"`
	// FileIDs llm FileIDs
	FileIDs []string `json:"file_id,omitempty"`
  // VideoURLs
  VideoURLs []string `json:"video_url,omitempty"`
}
