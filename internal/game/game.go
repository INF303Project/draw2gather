package game

import (
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/alexedwards/scs/v2"
)

type Game struct {
	id       string
	owner    string
	db       *firestore.Client
	sessions *scs.SessionManager

	ch         chan *Message
	register   chan *Player
	unregister chan *Player

	state   state
	stateCh chan state

	targetScore int
	words       map[string]struct{}
	players     map[string]*Player
	scores      map[*Player]int
	drawerQueue []*Player

	commands        []*Message
	currentWord     string
	currentDrawer   *Player
	answeredPlayers map[*Player]struct{}
}

type GameSetting struct {
	ID          string
	Owner       string
	TargetScore int
	Words       []string
	DB          *firestore.Client
	Sessions    *scs.SessionManager
}

func NewGame(settings *GameSetting) *Game {
	return &Game{
		id:       settings.ID,
		owner:    settings.Owner,
		db:       settings.DB,
		sessions: settings.Sessions,

		ch:         make(chan *Message),
		register:   make(chan *Player),
		unregister: make(chan *Player),

		state:   &waiting{},
		stateCh: make(chan state),

		targetScore: settings.TargetScore,
		words:       make(map[string]struct{}),
		players:     make(map[string]*Player),
		scores:      make(map[*Player]int),
		drawerQueue: []*Player{},

		commands:        []*Message{},
		currentWord:     "",
		currentDrawer:   nil,
		answeredPlayers: make(map[*Player]struct{}),
	}
}

func (g *Game) Register(p *Player) {
	g.register <- p
}

func (g *Game) broadcastMessage(m *Message) {
	for _, p := range g.players {
		p.ch <- m
	}
}

func (g *Game) Run() {
	for {
		select {
		case p := <-g.register:
			g.players[p.id] = p
			g.scores[p] = 0
			// Send state and commands to new player

		case p := <-g.unregister:
			g.sessions.Remove(p.ctx, "game_id")
			_, _, err := g.sessions.Commit(p.ctx)
			if err != nil {
				fmt.Println(err)
			}

			g.db.Collection("games").Doc(g.id).Update(p.ctx, []firestore.Update{
				{Path: "current_players", Value: firestore.Increment(-1)},
			})

			delete(g.players, p.id)
			delete(g.scores, p)
			close(p.ch)

		case msg := <-g.ch:
			state := g.state.HandleMessage(g, msg)
			if state != nil {
				g.state.Exit(g)
				g.state = state
				g.state.Enter(g)
			}

		case state := <-g.stateCh:
			g.state.Exit(g)
			g.state = state
			g.state.Enter(g)
		}
	}
}

func (g *Game) handleBoardAction(m *Message) error {
	if m.player != g.currentDrawer {
		return errors.New("not your turn")
	}
	if _, err := m.decodeMessage(); err != nil {
		return err
	}

	g.commands = append(g.commands, m)
	for _, p := range g.players {
		if p != m.player {
			p.ch <- m
		}
	}

	return nil
}

func (g *Game) handleStart(m *Message) error {
	if m.player.id != g.owner {
		return errors.New("not owner")
	}

	if len(g.players) < 2 {
		return errors.New("not enough players")
	}

	g.broadcastMessage(m)
	return nil
}

func (g *Game) handleQuit(m *Message) error {
	return nil
}

func (g *Game) handlePick(m *Message) error {
	return nil
}

func (g *Game) handleGuess(m *Message) error {
	return nil
}

func (g *Game) handleChat(m *Message) error {
	return nil
}

func (g *Game) handleKick(m *Message) error {
	return nil
}
