package main

import (
	"context"
	"log"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"github.com/alperenunal/draw2gather/internal/api"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/api/option"
)

func main() {
	opt := option.WithCredentialsFile("admin-sdk.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalln(err)
	}

	handler, err := api.NewHandler(app)
	if err != nil {
		log.Fatalln(err)
	}

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist("api.draw2gather.online"),
	}

	go func() {
		err := http.ListenAndServe(":80", certManager.HTTPHandler(nil))
		if err != nil {
			log.Fatalln(err)
		}
	}()

	server := &http.Server{
		Addr:      ":443",
		Handler:   handler,
		TLSConfig: certManager.TLSConfig(),
	}

	log.Fatalln(server.ListenAndServeTLS("", ""))
}
