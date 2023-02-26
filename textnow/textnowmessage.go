package textnow

import (
	"github.com/jckimble/smsprovider"
	"time"
)

type textNowMessage struct {
	contact     string
	message     string
	messagetype int
	read        bool
	id          string
	time        string
}

func (t textNowMessage) Attachments() []smsprovider.Attachment {
	if t.messagetype != 2 {
		return nil
	}
	return []smsprovider.Attachment{
		smsprovider.SMSAttachment{t.message},
	}
}

func (t textNowMessage) Message() string {
	if t.messagetype != 1 {
		return ""
	}
	return t.message
}
func (t textNowMessage) Source() string {
	return t.contact
}
func (t textNowMessage) Time() time.Time {
	ti, _ := time.Parse("2006-01-02T15:04:05Z", t.time)
	return ti
}
func (t textNowMessage) Read() bool {
	return t.read
}
