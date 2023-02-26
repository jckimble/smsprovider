package smsprovider_test

import (
	"github.com/jckimble/smsprovider"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSMSAttachment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	}))
	defer server.Close()
	attach := smsprovider.SMSAttachment{
		URL: server.URL + "/test.png",
	}
	if attach.GetMimeType() != "image/png" {
		t.Errorf("Types do not match %s != %s", attach.GetMimeType(), "image/png")
	}
	r, err := attach.GetReader()
	if err != nil {
		t.Errorf("Failed Reader Returned: %s", err.Error())
	}
	if r != nil {
		r.Close()
	}
}
