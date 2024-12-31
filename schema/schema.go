package schema

// Schema is message schema interface
type Schema interface {
	// String returns a string representation of the schema
	String() string
	// Snapshot returns a snapshot of the schema
	Snapshot() Schema
	// Attachement() returns schema attchement
	Attachement() *Attachement
}
