package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.com/jckimble/smsprovider"
	"gitlab.com/jckimble/smsprovider/signal"
	"gitlab.com/jckimble/smsprovider/textnow"
	"log"
	"net/http"
	"os"
	ossignal "os/signal"
	"strings"
	"sync"
	"time"
)

//Note: This is an example, generally you would want to add
//		a middleware to authenticate request where the api
//		couldn't be abused!
func main() {
	textnow := &textnow.TextNow{
		Username: "username",
		Password: "password",
	}
	signalsms := &signal.Signal{
		GetPhoneNumberFunc: textnow.GetPhoneNumber,
		StorageDir:         "./.signal",
	}
	signalsms.Handler = func(m smsprovider.Message) {
	}
	textnow.Handler = func(m smsprovider.Message) {
		if strings.HasPrefix(m.Message(), "Your Signal verification code:") {
			spl := strings.Split(m.Message(), ":")
			code := strings.TrimSpace(spl[1])
			signalsms.SetVerificationCode(code)
		}
		textnow.DeleteMessage(m)
	}
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		if err := textnow.Setup(); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := signalsms.Setup(); err != nil {
			log.Fatal(err)
		}
	}()
	r := mux.NewRouter()
	r.HandleFunc("/sms", SendFunc(textnow)).Methods("POST")
	r.HandleFunc("/signal", SendFunc(signalsms)).Methods("POST")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		defer wg.Done()
		c := make(chan os.Signal, 1)
		ossignal.Notify(c, os.Interrupt)
		<-c
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		srv.Shutdown(ctx)
		signalsms.Shutdown()
		textnow.Shutdown()
	}()
	wg.Wait()
}
func SendFunc(p smsprovider.Provider) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("contact") == "" || (r.Header.Get("Content-Type") != "multipart/form-data" && r.FormValue("message") == "") {
			res := APIResponse{
				Code:    400,
				Message: "Missing Required Fields",
			}
			if err := res.Write(w); err != nil {
				log.Printf("%s", err)
			}
			return
		}
		if r.Header.Get("Content-Type") == "multipart/form-data" { //attachment
			if ap, ok := interface{}(p).(smsprovider.AttachmentProvider); ok {
				file, _, err := r.FormFile("attachment")
				if err != nil {
					log.Printf("%s", err)
					res := APIResponse{
						Code:    500,
						Message: "Unable to send Attachment",
					}
					if err := res.Write(w); err != nil {
						log.Printf("%s", err)
					}
					return
				}
				defer file.Close()
				if err := ap.SendAttachment(r.FormValue("contact"), r.FormValue("message"), file); err != nil {
					log.Printf("%s", err)
					res := APIResponse{
						Code:    500,
						Message: "Unable to send Attachment",
					}
					if err := res.Write(w); err != nil {
						log.Printf("%s", err)
					}
					return
				}
				res := APIResponse{
					Code:    200,
					Message: "Attachment Sent",
				}
				if err := res.Write(w); err != nil {
					log.Printf("%s", err)
				}
			} else {
				res := APIResponse{
					Code:    403,
					Message: "Provider doesn't support attachments",
				}
				if err := res.Write(w); err != nil {
					log.Printf("%s", err)
				}
			}
		} else {
			if err := p.SendMessage(r.FormValue("contact"), r.FormValue("message")); err != nil {
				log.Printf("%s", err)
				res := APIResponse{
					Code:    500,
					Message: "Unable to send Message",
				}
				if err := res.Write(w); err != nil {
					log.Printf("%s", err)
				}
				return
			}
			res := APIResponse{
				Code:    200,
				Message: "Message Sent",
			}
			if err := res.Write(w); err != nil {
				log.Printf("%s", err)
			}
		}
	}
}

type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func (ar APIResponse) Write(w http.ResponseWriter) error {
	w.WriteHeader(ar.Code)
	return json.NewEncoder(w).Encode(ar)
}
