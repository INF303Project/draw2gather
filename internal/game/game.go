package game

import (
	"cloud.google.com/go/firestore"
)

type Game struct {
	id    string
	owner string
	db    *firestore.Client

	ch         chan *Message
	register   chan *Player
	unregister chan *Player

	state   state
	stateCh chan state

	players any
	scores  any
	words   any

	commands      []string
	currentWord   string
	currentPlayer string
}

func NewGame(id string, owner string, db *firestore.Client) *Game {
	return &Game{
		id:    id,
		owner: owner,
		db:    db,

		ch:         make(chan *Message),
		register:   make(chan *Player),
		unregister: make(chan *Player),

		players: nil,
		scores:  nil,
		words:   nil,

		state:   nil,
		stateCh: make(chan state),

		commands:      []string{},
		currentWord:   "",
		currentPlayer: "",
	}
}

func (g *Game) Register(p *Player) {
	g.register <- p
}

func (g *Game) Run() {
	for {
		select {}
	}
}
