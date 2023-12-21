package game

import (
	"context"
	"encoding/json"
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

	ch       chan *Message
	register chan *Player

	state   state
	stateCh chan state

	targetScore int
	words       map[string]struct{}
	players     map[string]*Player
	drawerQueue []*Player

	commands        []*Message
	currentWord     string
	currentDrawer   *Player
	answeredPlayers map[*Player]struct{}
}

type GameSettings struct {
	ID          string
	Owner       string
	TargetScore int
	Words       []string
	DB          *firestore.Client
	Sessions    *scs.SessionManager
}

func NewGame(settings *GameSettings) *Game {
	return &Game{
		id:       settings.ID,
		owner:    settings.Owner,
		db:       settings.DB,
		sessions: settings.Sessions,

		ch:       make(chan *Message),
		register: make(chan *Player),

		state:   &waitingState{},
		stateCh: make(chan state),

		targetScore: settings.TargetScore,
		words:       make(map[string]struct{}),
		players:     make(map[string]*Player),
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

func (g *Game) Run() {
	for {
		select {
		case p := <-g.register:
			g.handleJoin(p)

		case msg, ok := <-g.ch:
			if !ok {
				g.close()
				return
			}

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

func (g *Game) close() {
	Hub.Delete(g.id)
	g.db.Collection("games").Doc(g.id).Delete(context.Background())
}

func (g *Game) sendToPlayer(p *Player, m *Message) {
	p.ch <- m
}

func (g *Game) sendExceptPlayer(p *Player, m *Message) {
	for _, p := range g.players {
		if p != m.player {
			p.ch <- m
		}
	}
}

func (g *Game) sendToAll(m *Message) {
	for _, p := range g.players {
		p.ch <- m
	}
}

func (g *Game) pickPlayer() *Player {
	p := g.drawerQueue[0]
	g.drawerQueue = g.drawerQueue[1:]
	g.drawerQueue = append(g.drawerQueue, p)
	return p
}

func (g *Game) pickWords() [2]string {
	var words [2]string
	for i := 0; i < 2; i++ {
		for word := range g.words {
			words[i] = word
			delete(g.words, word)
			break
		}
	}
	return words
}

func (g *Game) handleJoin(p *Player) {
	g.players[p.ID] = p
	// Send state to player
}

func (g *Game) handleQuit(m *Message) (state, error) {
	g.sessions.Remove(m.player.ctx, "game_id")
	_, _, err := g.sessions.Commit(m.player.ctx)
	if err != nil {
		fmt.Println(err)
	}

	delete(g.players, m.player.ID)
	close(m.player.ch)

	if len(g.players) == 0 {
		close(g.ch)
		return nil, nil
	}

	g.db.Collection("games").Doc(g.id).Update(context.Background(), []firestore.Update{
		{Path: "current_players", Value: firestore.Increment(-1)},
	})

	owner := ""
	if m.player.ID == g.owner {
		for id := range g.players {
			owner = id
			break
		}
	}

	msg := map[string]any{
		"player": m.player.ID,
	}
	payload, _ := json.Marshal(msg)
	g.sendToAll(&Message{
		Action:  quit,
		Payload: string(payload),
	})

	msg = map[string]any{
		"owner": owner,
	}
	payload, _ = json.Marshal(msg)
	g.sendToAll(&Message{
		// Action:  ownerChange,
		Payload: string(payload),
	})

	return nil, nil
}

func (g *Game) handleKick(m *Message) (state, error) {
	if m.player.ID != g.owner {
		return nil, errors.New("not owner")
	}

	in, err := m.decodeMessage()
	if err != nil {
		return nil, err
	}
	p := in.(payload[string])

	if _, ok := g.players[p.Value]; !ok {
		return nil, errors.New("player not found")
	}

	delete(g.players, p.Value)
	close(g.players[p.Value].ch)

	msg := userPayload[string]{
		ID: p.Value,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	g.sendToAll(&Message{
		Action:  quit,
		Payload: string(payload),
	})

	return nil, nil
}

func (g *Game) handleChat(m *Message) error {
	in, err := m.decodeMessage()
	if err != nil {
		return err
	}
	p := in.(payload[string])

	msg := userPayload[string]{
		ID:    m.player.ID,
		Value: p.Value,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	g.sendToAll(&Message{
		Action:  chat,
		Payload: string(payload),
	})

	return nil
}

func (g *Game) handleStart(m *Message) (state, error) {
	if m.player.ID != g.owner {
		return nil, errors.New("not owner")
	}

	if len(g.players) < 2 {
		return nil, errors.New("not enough players")
	}

	return &startingState{}, nil
}

func (g *Game) handlePick(m *Message) (state, error) {
	if m.player != g.currentDrawer {
		return nil, errors.New("not your turn")
	}

	in, err := m.decodeMessage()
	if err != nil {
		return nil, err
	}
	p := in.(payload[string])

	g.currentWord = p.Value

	return nil, nil
}

func (g *Game) handleGuess(m *Message) (state, error) {
	return nil, nil
}

func (g *Game) handleBoardAction(m *Message) error {
	if m.player != g.currentDrawer {
		return errors.New("not your turn")
	}
	if _, err := m.decodeMessage(); err != nil {
		return err
	}

	g.commands = append(g.commands, m)
	g.sendExceptPlayer(m.player, m)

	return nil
}
