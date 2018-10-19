package textnow

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"gitlab.com/jckimble/smsprovider"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	TN_API_URL    = "https://api.textnow.me/api2.0/"
	TN_SIGN_KEY   = "f8ab2ceca9163724b6d126aea9620339"
	TN_USER_AGENT = "TextNow 5.15.0 (Android SDK built for x86; Android OS 5.1.1; en_US)"
)

// TextNow is the provider for TextNow standard sms, it is a US only service
type TextNow struct {
	Username string
	Password string
	Handler  func(smsprovider.Message)
	Delay    time.Duration

	loginid  string
	username string
	userInfo map[string]interface{}
	done     chan bool
}

func (t TextNow) genSignature(method, node, query string) string {
	text := fmt.Sprintf("%s%s%s%s", TN_SIGN_KEY, method, node, query)
	return fmt.Sprintf("%x", md5.Sum([]byte(text)))
}

func (t TextNow) getUserInfo() (map[string]interface{}, error) {
	node := "users/" + t.username
	query := "?client_type=TN_ANDROID&client_id=" + t.loginid
	sign := t.genSignature("GET", node, query)
	return t.sendReq("GET", node, query, sign, nil)
}
func (t TextNow) logout() error {
	node := "sessions"
	query := "?client_type=TN_ANDROID&client_id=" + t.loginid
	sign := t.genSignature("DELETE", node, query)
	_, err := t.sendReq("DELETE", node, query, sign, nil)
	return err
}

// Trigger an graceful shutdown
func (t *TextNow) Shutdown() error {
	t.done <- true
	return t.logout()
}

// Setup TextNow and wait for incoming messages
func (t *TextNow) Setup() error {
	if t.loginid == "" {
		if err := t.login(); err != nil {
			return err
		}
	}
	if t.Delay == 0 {
		t.Delay = 30 * time.Second
	}
	if t.Handler == nil {
		return nil
	}
	t.done = make(chan bool, 1)
	ticker := time.NewTicker(t.Delay)
	for {
		select {
		case <-t.done:
			return nil
		case <-ticker.C:
			node := "users/" + t.username + "/messages"
			query := "?client_type=TN_ANDROID&client_id=" + t.loginid + "&get_all=1&page_size=30&start_message_id=1"
			sign := t.genSignature("GET", node, query)
			req, err := t.sendReq("GET", node, query, sign, nil)
			if err != nil {
				return err
			}
			if req["messages"] == nil {
				return fmt.Errorf("Account Logged Out")
			}
			for _, msg := range req["messages"].([]interface{}) {
				m := msg.(map[string]interface{})
				tnm := textNowMessage{
					contact:     m["contact_value"].(string),
					message:     m["message"].(string),
					messagetype: int(m["message_type"].(float64)),
					read:        m["read"].(bool),
					id:          strconv.Itoa(int(m["id"].(float64))),
					time:        m["date"].(string),
				}
				t.Handler(tnm)
			}
		}
	}
}

// DeleteMessage deletes a message from the TextNow's servers.
func (t TextNow) DeleteMessage(m smsprovider.Message) error {
	tnm, ok := m.(textNowMessage)
	if !ok {
		return fmt.Errorf("Message is not an TextNow Message")
	}
	node := "users/" + t.username + "/conversations/" + tnm.contact
	query := "?client_type=TN_ANDROID&client_id=" + t.loginid + "&latest_message_id=" + tnm.id
	data := map[string]bool{
		"read": true,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	jsonData2, err := json.Marshal(map[string]interface{}{
		"json": string(jsonData),
	})
	if err != nil {
		return err
	}
	sign := t.genSignature("DELETE", node, query+string(jsonData))
	_, err = t.sendReq("DELETE", node, query, sign, bytes.NewReader(jsonData2))
	if err != nil {
		return err
	}
	return nil
}

// MarkRead marks a message as read on TextNow's servers
func (t TextNow) MarkRead(m smsprovider.Message) error {
	tnm, ok := m.(textNowMessage)
	if !ok {
		return fmt.Errorf("Message is not an TextNow Message")
	}
	node := "users/" + t.username + "/conversations/" + tnm.contact
	query := "?client_type=TN_ANDROID&client_id=" + t.loginid + "&latest_message_id=" + tnm.id
	data := map[string]bool{
		"read": true,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	jsonData2, err := json.Marshal(map[string]interface{}{
		"json": string(jsonData),
	})
	if err != nil {
		return err
	}
	sign := t.genSignature("PATCH", node, query+string(jsonData))
	_, err = t.sendReq("PATCH", node, query, sign, bytes.NewReader(jsonData2))
	if err != nil {
		return err
	}
	return nil
}

// SendMessage sends an message to a phone number
func (t *TextNow) SendMessage(to, msg string) error {
	if t.loginid == "" {
		if err := t.login(); err != nil {
			return err
		}
	}
	node := "users/" + t.username + "/messages"
	query := "?client_type=TN_ANDROID&client_id=" + t.loginid
	data := map[string]string{
		"from_name":     t.userInfo["first_name"].(string),
		"contact_type":  "2",
		"contact_value": "+" + to,
		"message":       msg,
		"to_name":       "",
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	jsonData2, err := json.Marshal(map[string]interface{}{
		"json": data,
	})
	if err != nil {
		return err
	}
	sign := t.genSignature("POST", node, query+string(jsonData))
	_, err = t.sendReq("POST", node, query, sign, bytes.NewReader(jsonData2))
	if err != nil {
		return err
	}
	return nil
}

func (t *TextNow) login() error {
	loginData := map[string]string{
		"password":    t.Password,
		"username":    t.Username,
		"esn":         "00000000000000",
		"os_version":  "22",
		"app_version": "5.15.0",
		"iccid":       "89014103211118510720",
	}
	node := "sessions"
	query := "?client_type=TN_ANDROID"
	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return err
	}
	jsonData2, err := json.Marshal(map[string]interface{}{
		"json": loginData,
	})
	if err != nil {
		return err
	}
	sign := t.genSignature("POST", node, query+string(jsonData))
	req, err := t.sendReq("POST", node, query, sign, bytes.NewReader(jsonData2))
	if err != nil {
		return err
	}
	if req["id"] != nil {
		t.loginid = req["id"].(string)
		t.username = req["username"].(string)
		t.userInfo, err = t.getUserInfo()
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("Unable to login")
}

// GetPhoneNumber returns the number given by TextNow
func (t *TextNow) GetPhoneNumber() (string, error) {
	if t.loginid == "" {
		if err := t.login(); err != nil {
			return "", err
		}
	}
	return "+1" + t.userInfo["phone_number"].(string), nil //US Only
}

func (t TextNow) sendReq(method, node, query, sign string, data io.Reader) (map[string]interface{}, error) {
	if method != "POST" && method != "GET" && method != "PATCH" && method != "DELETE" {
		return nil, fmt.Errorf("Unsupported Method: %s", method)
	}
	url := fmt.Sprintf("%s%s%s&signature=%s", TN_API_URL, node, query, sign)
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", TN_USER_AGENT)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	m := map[string]interface{}{}
	if method != "PATCH" && method != "DELETE" {
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&m); err != nil && err.Error() != "EOF" {
			return nil, err
		}
	}
	return m, nil
}
