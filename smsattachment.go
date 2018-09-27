package smsprovider

import (
	"io"
	"net/http"

	"mime"
	"path/filepath"
)

// SMSAttachment is an struct for sms attachments, standard sms sends
// an url for an attachment. This is here as an convenience struct for
// new sms providers.
type SMSAttachment struct {
	URL string
}

func (a SMSAttachment) GetMimeType() string {
	if ext := filepath.Ext(a.URL); ext != "" {
		return mime.TypeByExtension(ext)
	}
	return ""
}

func (a SMSAttachment) GetReader() (io.ReadCloser, error) {
	resp, err := http.Get(a.URL)
	if err != nil {
		return nil, err
	}
	return resp.Body, err
}
