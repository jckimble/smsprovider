package smsprovider

import (
	"io"
)

// AttachmentProvider is a provider type that shows that you can send
// attachments from it, it is not required but recommended.
type AttachmentProvider interface {
	SendAttachment(string, string, io.Reader) error
}

// All Providers must implement this interface, it has the required functions
// for a provider to be useful
type Provider interface {
	SendMessage(string, string) error
	DeleteMessage(Message) error
	GetPhoneNumber() (string, error)

	Setup() error
	Shutdown() error
}
