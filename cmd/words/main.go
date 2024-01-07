package main

import (
	"bufio"
	"context"
	"os"
	"slices"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	_ "github.com/joho/godotenv/autoload"
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

	store(firestore, "TR", trFile)
	store(firestore, "EN", enFile)
	store(firestore, "DE", deFile)
}

func store(firestore *firestore.Client, language, fileName string) {
	var words []string
	tr, _ := os.Open(fileName)
	scanner := bufio.NewScanner(tr)

	for scanner.Scan() {
		word := scanner.Text()
		words = append(words, strings.ToLower(word))
	}
	tr.Close()

	slices.Sort(words)
	ctx := context.Background()
	firestore.Collection("word_sets").Doc(language).
		Set(ctx, &WordSet{
			Lang:  language,
			Name:  "default",
			Words: words,
		})
}
