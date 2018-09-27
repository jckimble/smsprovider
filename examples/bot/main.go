package main

import (
	"gitlab.com/jckimble/smsprovider"
	"gitlab.com/jckimble/smsprovider/signal"
	"gitlab.com/jckimble/smsprovider/textnow"
	"log"
	"os"
	ossignal "os/signal"
	"strings"
	"sync"
)

func main() {
	textnow := &textnow.TextNow{
		Username: "username",
		Password: "password",
	}
	signalsms := &signal.Signal{
		GetPhoneNumberFunc: textnow.GetPhoneNumber,
		StorageDir:         "./.signal",
	}
	signalsms.Handler = echoHandler(signalsms)
	textnow.Handler = func(m smsprovider.Message) {
		if strings.HasPrefix(m.Message(), "Your Signal verification code:") {
			spl := strings.Split(m.Message(), ":")
			code := strings.TrimSpace(spl[1])
			signalsms.SetVerificationCode(code)
		} else {
			echoHandler(textnow)(m)
		}
		textnow.DeleteMessage(m)
	}
	var wg sync.WaitGroup
	wg.Add(3)
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
	go func() {
		defer wg.Done()
		c := make(chan os.Signal, 1)
		ossignal.Notify(c, os.Interrupt)
		<-c
		signalsms.Shutdown()
		textnow.Shutdown()
	}()
	wg.Wait()
}
func echoHandler(p smsprovider.Provider) func(smsprovider.Message) {
	return func(m smsprovider.Message) {
		p.SendMessage(m.Source(), m.Message())
	}
}
