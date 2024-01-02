package main

import (
	"context"
	"log"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"github.com/alperenunal/draw2gather/internal/api"
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

	log.Println("Listening on :8080")
	log.Fatalln(http.ListenAndServe(":8080", handler))
}
