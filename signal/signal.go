package signal

import (
	"github.com/aebruno/textsecure"
	"github.com/jckimble/smsprovider"
	"io"
	"log"
	"strings"
	"time"
)

// Signal is the provider for signal sms
type Signal struct {
	Handler            func(smsprovider.Message)
	GetPhoneNumberFunc func() (string, error)
	StorageDir         string

	verificationCode string
}

// SendMessage sends an message to a group or signal user.
func (s Signal) SendMessage(to string, msg string) error {
	to = strings.TrimSpace(to)
	if len(to) == 4 {
		_, err := textsecure.SendGroupMessage(to, msg)
		return err
	} else {
		if !strings.HasPrefix(to, "+") {
			to = "+" + to
		}
		_, err := textsecure.SendMessage(to, msg)
		return err
	}
}

// DeleteMessage is an noop function in signal since messages are not stored on a remote server
func (s Signal) DeleteMessage(m smsprovider.Message) error {
	return nil
}

// Trigger an graceful shutdown
func (s Signal) Shutdown() error {
	return textsecure.StopListening()
}

// Setup signal and wait for incoming messages
func (s *Signal) Setup() error {
	client := &textsecure.Client{
		GetConfig: func() (*textsecure.Config, error) {
			num, err := s.GetPhoneNumberFunc()
			if err != nil {
				return nil, err
			}
			return &textsecure.Config{
				Tel:                num,
				Server:             "https://textsecure-service.whispersystems.org:443",
				VerificationType:   "sms",
				StorageDir:         s.StorageDir,
				UnencryptedStorage: true,
				AlwaysTrustPeerID:  true,
			}, nil
		},
		GetVerificationCode: func() string {
			for {
				if s.verificationCode != "" {
					return s.verificationCode
				}
				time.Sleep(time.Second)
			}
		},
		RegistrationDone: func() {
			log.Printf("Registration Done\n")
		},
		MessageHandler: func(m *textsecure.Message) {
			if s.Handler != nil {
				s.Handler(signalMessage{m})
			}
		},
	}
	if err := textsecure.Setup(client); err != nil {
		return err
	}
	if err := textsecure.StartListening(); err != nil {
		return err
	}
	return nil
}

// SendAttachment sends an attachment to a group or signal user.
func (s Signal) SendAttachment(to string, msg string, f io.Reader) error {
	to = strings.TrimSpace(to)
	if len(to) == 4 {
		_, err := textsecure.SendGroupAttachment(to, msg, f)
		return err
	} else {
		if !strings.HasPrefix(to, "+") {
			to = "+" + to
		}
		_, err := textsecure.SendAttachment(to, msg, f)
		return err
	}
}

// GetPhoneNumber returns the number given by GetPhoneNumberFunc
func (s Signal) GetPhoneNumber() (string, error) {
	return s.GetPhoneNumberFunc()
}

// SetVerificationCode sets the code required on signal registration
func (s *Signal) SetVerificationCode(code string) {
	s.verificationCode = code
}
