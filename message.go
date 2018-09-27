package smsprovider

import (
	"io"
	"time"
)

// Message is a central interface messages from providers should implement
// it provides everything required from a message
type Message interface {
	Source() string
	Message() string
	Attachments() []Attachment
	Time() time.Time
	Read() bool
}

// Attachment is a central interface to get attachments from a message
type Attachment interface {
	GetMimeType() string
	GetReader() (io.ReadCloser, error)
}
