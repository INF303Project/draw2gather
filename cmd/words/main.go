package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"slices"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type WordSet struct {
	Lang  string   `firestore:"language" json:"language"`
	Name  string   `firestore:"name" json:"name"`
	Words []string `firestore:"words" json:"words"`
}

const (
	trFile = "./cmd/words/tr.txt"
	enFile = "./cmd/words/en.txt"
	deFile = "./cmd/words/de.txt"
)

func main() {
	app, err := firebase.NewApp(context.Background(), nil,
		option.WithCredentialsFile("admin-sdk.json"))
	if err != nil {
		panic(err)
	}

	firestore, err := app.Firestore(context.Background())
	if err != nil {
		panic(err)
	}

	storeTR(firestore)
}

func storeTR(firestore *firestore.Client) {
	var words []string
	tr, _ := os.Open(trFile)
	scanner := bufio.NewScanner(tr)

	for scanner.Scan() {
		word := scanner.Text()
		words = append(words, strings.ToLower(word))
	}
	tr.Close()

	slices.Sort(words)
	ctx := context.Background()
	firestore.Collection("word_sets").Doc("TR").
		Set(ctx, &WordSet{
			Lang:  "TR",
			Name:  "default",
			Words: words,
		})
}

func sortTR() {
	var words []string
	tr, _ := os.Open(trFile)
	scanner := bufio.NewScanner(tr)

	for scanner.Scan() {
		word := scanner.Text()
		words = append(words, strings.ToLower(word))
	}
	tr.Close()

	slices.Sort(words)
	tr, _ = os.OpenFile(trFile, os.O_WRONLY|os.O_TRUNC, 0644)
	io.WriteString(tr, strings.Join(words, "\n"))
	tr.Close()
}
