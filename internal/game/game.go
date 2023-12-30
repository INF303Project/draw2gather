package game

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/alexedwards/scs/v2"
)

type Game struct {
	id       string
	owner    string
	db       *firestore.Client
	sessions *scs.SessionManager

	closed   bool
	ch       chan *Message
	register chan *Player

	state   state
	stateCh chan state

	targetScore int
	words       map[string]struct{}
	dictionary  map[string]struct{}
	players     map[string]*Player
	playerQueue []*Player

	currentPlayer   *Player
	currentWord     string
	commands        []*Message
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
	words := make(map[string]struct{})
	for _, word := range settings.Words {
		word = strings.ToLower(word)
		words[word] = struct{}{}
	}

	return &Game{
		id:       settings.ID,
		owner:    settings.Owner,
		db:       settings.DB,
		sessions: settings.Sessions,

		closed:   false,
		ch:       make(chan *Message),
		register: make(chan *Player),

		state:   &waitingState{},
		stateCh: make(chan state),

		targetScore: settings.TargetScore,
		dictionary:  words,
		words:       words,
		players:     make(map[string]*Player),
		playerQueue: []*Player{},

		currentPlayer:   nil,
		currentWord:     "",
		commands:        []*Message{},
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

			slog.Info("Received message", slog.Int("action", int(msg.Action)))
			state := g.state.HandleMessage(g, msg)
			if state != nil {
				g.state.Exit(g)
				g.state = state
				g.state.Enter(g)
			}

		case state := <-g.stateCh:
			if state != nil {
				g.state.Exit(g)
				g.state = state
				g.state.Enter(g)
			}
		}
	}
}

func (g *Game) Closed() bool {
	return g.closed
}

func (g *Game) close() {
	g.closed = true
	Hub.Delete(g.id)
	g.db.Collection("games").Doc(g.id).Delete(context.Background())
}

func (g *Game) sendToPlayer(p *Player, m *Message) {
	p.ch <- m
}

func (g *Game) sendExceptPlayer(player *Player, m *Message) {
	for _, p := range g.players {
		if p != player {
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
	p := g.playerQueue[0]
	g.playerQueue = g.playerQueue[1:]
	g.playerQueue = append(g.playerQueue, p)
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
	g.words[words[0]] = struct{}{}
	g.words[words[1]] = struct{}{}
	return words
}

func (g *Game) getNextPoint() int {
	return 10 - len(g.answeredPlayers)
}

func (g *Game) removePlayer(player *Player) (state, error) {
	if _, ok := g.players[player.ID]; !ok {
		return nil, errors.New("player not found")
	}

	g.sessions.Remove(player.ctx, "game_id")
	g.sessions.Commit(player.ctx)

	delete(g.players, player.ID)
	close(player.ch)
	g.playerQueue = slices.DeleteFunc(g.playerQueue, func(p *Player) bool {
		return p == player
	})

	if len(g.players) == 0 {
		return &closingState{}, nil
	}

	g.db.Collection("games").Doc(g.id).Update(context.Background(), []firestore.Update{
		{
			Path:  "current_players",
			Value: firestore.ArrayRemove(g.sessions.GetString(player.ctx, "name")),
		},
	})

	msg := newMessage(quit, player.ID)
	g.sendToAll(msg)

	if player.ID == g.owner {
		owner := ""
		for id := range g.players {
			owner = id
			break
		}

		g.owner = owner
		msg := newMessage(newOwner, owner)
		g.sendToAll(msg)
	}

	if len(g.players) == 1 {
		return &waitingState{}, nil
	}
	if g.currentPlayer == player {
		return &startingState{}, nil
	}

	return nil, nil
}

func (g *Game) handleJoin(p *Player) {
	if _, ok := g.players[p.ID]; ok {
		return
	}

	g.players[p.ID] = p
	g.playerQueue = append(g.playerQueue, p)

	var greetPayload gamePayload

	switch g.state.(type) {
	case *waitingState:
		greetPayload.State = "waiting"
	case *startingState:
		greetPayload.State = "starting"
	case *pickingState:
		greetPayload.State = "picking"
	case *drawingState:
		greetPayload.State = "drawing"
	case *endingState:
		greetPayload.State = "ending"
	default:
		return
	}

	greetPayload.Player = p.ID
	greetPayload.Owner = g.owner
	if g.currentPlayer != nil {
		greetPayload.CurrentPlayer = g.currentPlayer.ID
	}
	greetPayload.Players = g.players
	greetPayload.Commands = g.commands

	msg := newMessage(greet, greetPayload)
	g.sendToPlayer(p, msg)

	msg = newMessage(join, p)
	g.sendExceptPlayer(p, msg)
}

func (g *Game) handleQuit(m *Message) (state, error) {
	return g.removePlayer(m.player)
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

	if p.Value == g.owner {
		return nil, errors.New("cannot kick owner")
	}

	if player, ok := g.players[p.Value]; !ok {
		return nil, errors.New("player not found")
	} else {
		g.db.Collection("games").Doc(g.id).Update(context.Background(), []firestore.Update{
			{Path: "banned_players", Value: firestore.ArrayUnion(player.ID)},
		})
		g.sendToPlayer(player, newEmptyMessage(kick))
		return g.removePlayer(player)
	}
}

func (g *Game) handleChat(m *Message) error {
	in, err := m.decodeMessage()
	if err != nil {
		return err
	}
	p := in.(payload[string])

	msg := newMessage(chat, messagePayload{
		Player:  m.player.ID,
		Message: p.Value,
	})
	g.sendToAll(msg)

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
	if m.player != g.currentPlayer {
		return nil, errors.New("not your turn")
	}

	in, err := m.decodeMessage()
	if err != nil {
		return nil, err
	}
	p := in.(payload[string])

	g.currentWord = p.Value
	delete(g.words, p.Value)

	return &drawingState{}, nil
}

func (g *Game) handleBoardAction(m *Message) error {
	if m.player != g.currentPlayer {
		return errors.New("not your turn")
	}
	if _, err := m.decodeMessage(); err != nil {
		return err
	}

	g.commands = append(g.commands, m)
	g.sendExceptPlayer(m.player, m)

	return nil
}

func (g *Game) handleGuess(m *Message) (state, error) {
	if _, ok := g.answeredPlayers[m.player]; ok {
		return nil, errors.New("already answered")
	}

	in, err := m.decodeMessage()
	if err != nil {
		return nil, err
	}
	p := in.(payload[string])

	if strings.ToLower(p.Value) == g.currentWord {
		msg := newMessage(correctGuess, m.player.ID)
		g.sendToAll(msg)

		if len(g.answeredPlayers) == 0 {
			g.currentPlayer.Score += 10
			msg := newMessage(updateScore, scorePayload{
				Player: g.currentPlayer.ID,
				Score:  g.currentPlayer.Score,
			})
			g.sendToAll(msg)
		}

		m.player.Score += g.getNextPoint()
		msg = newMessage(updateScore, scorePayload{
			Player: m.player.ID,
			Score:  m.player.Score,
		})
		g.sendToAll(msg)

		if g.currentPlayer.Score >= g.targetScore || m.player.Score >= g.targetScore {
			return &endingState{}, nil
		}

		g.answeredPlayers[m.player] = struct{}{}
		finished := true
		for _, player := range g.players {
			if player == g.currentPlayer {
				continue
			}
			if _, ok := g.answeredPlayers[player]; !ok {
				finished = false
				break
			}
		}

		if finished {
			g.currentPlayer.Score += 1
			msg := newMessage(updateScore, scorePayload{
				Player: g.currentPlayer.ID,
				Score:  g.currentPlayer.Score,
			})
			g.sendToAll(msg)

			if g.currentPlayer.Score >= g.targetScore {
				return &endingState{}, nil
			}

			return &startingState{}, nil
		}
	} else {
		msg := newMessage(guess, messagePayload{
			Player:  m.player.Name,
			Message: p.Value,
		})
		g.sendToAll(msg)
	}

	return nil, nil
}
