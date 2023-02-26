package signal

import (
	"github.com/aebruno/textsecure"
	"github.com/jckimble/smsprovider"
	"io"
	"io/ioutil"
	"time"
)

type signalMessage struct {
	Msg *textsecure.Message
}

func (m signalMessage) Attachments() []smsprovider.Attachment {
	ments := []smsprovider.Attachment{}
	for _, attach := range m.Msg.Attachments() {
		ments = append(ments, signalAttachment{attach})
	}
	return ments
}
func (m signalMessage) Message() string {
	return m.Msg.Message()
}
func (m signalMessage) Read() bool {
	return false
}
func (m signalMessage) Group() *textsecure.Group {
	return m.Msg.Group()
}
func (m signalMessage) Source() string {
	return m.Msg.Source()
}
func (m signalMessage) Time() time.Time {
	return time.Unix(int64(m.Msg.Timestamp()), 0)
}

type signalAttachment struct {
	Attach *textsecure.Attachment
}

func (a signalAttachment) GetMimeType() string {
	return a.Attach.MimeType
}
func (a signalAttachment) GetReader() (io.ReadCloser, error) {
	return ioutil.NopCloser(a.Attach.R), nil
}
