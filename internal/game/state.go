package game

import (
	"log/slog"
	"time"
)

type state interface {
	Enter(*Game)
	Exit(*Game)
	HandleMessage(*Game, *Message) state
}

type waitingState struct {
}

func (s *waitingState) Enter(g *Game) {
	slog.Info("Entering waiting state")
	clear(g.commands)
	clear(g.answeredPlayers)
	g.words = g.dictionary
	g.currentPlayer = nil
	g.currentWord = ""
	for _, p := range g.players {
		p.Score = 0
	}
	g.sendToAll(newEmptyMessage(waiting))
}

func (s *waitingState) Exit(g *Game) {
	slog.Info("Exiting waiting state")
}

func (s *waitingState) HandleMessage(g *Game, m *Message) state {
	switch m.Action {
	case quit:
		state, err := g.handleQuit(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	case kick:
		state, err := g.handleKick(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	case chat:
		err := g.handleChat(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return nil
	case start:
		state, err := g.handleStart(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	default:
		return nil
	}
}

type startingState struct {
	timer *time.Timer
}

func (s *startingState) Enter(g *Game) {
	slog.Info("Entering starting state")

	clear(g.commands)
	clear(g.answeredPlayers)
	g.currentPlayer = g.pickPlayer()

	msg := newMessage(starting, g.currentPlayer.ID)
	g.sendToAll(msg)

	s.timer = time.AfterFunc(5*time.Second, func() {
		g.stateCh <- &pickingState{}
	})
}

func (s *startingState) Exit(g *Game) {
	s.timer.Stop()
	slog.Info("Exiting starting state")
}

func (s *startingState) HandleMessage(g *Game, m *Message) state {
	switch m.Action {
	case quit:
		state, err := g.handleQuit(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	case kick:
		state, err := g.handleKick(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	case chat:
		err := g.handleChat(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return nil
	default:
		return nil
	}
}

type pickingState struct {
	timer *time.Timer
}

func (s *pickingState) Enter(g *Game) {
	slog.Info("Entering picking state")

	words := g.pickWords()

	msg := newEmptyMessage(picking)
	g.sendToAll(msg)

	msg = newMessage(pick, words)
	g.sendToPlayer(g.currentPlayer, msg)

	s.timer = time.AfterFunc(10*time.Second, func() {
		g.stateCh <- &startingState{}
	})
}

func (s *pickingState) Exit(g *Game) {
	s.timer.Stop()
	slog.Info("Exiting picking state")
}

func (s *pickingState) HandleMessage(g *Game, m *Message) state {
	switch m.Action {
	case quit:
		state, err := g.handleQuit(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	case kick:
		state, err := g.handleKick(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	case chat:
		err := g.handleChat(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return nil
	case pick:
		state, err := g.handlePick(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	}
	return nil
}

type drawingState struct {
	timer *time.Timer
}

func (s *drawingState) Enter(g *Game) {
	slog.Info("Entering drawing state")

	msg := newEmptyMessage(drawing)
	g.sendToAll(msg)

	s.timer = time.AfterFunc(1*time.Minute, func() {
		g.stateCh <- &startingState{}
	})
}

func (s *drawingState) Exit(g *Game) {
	s.timer.Stop()
	slog.Info("Exiting drawing state")
}

func (s *drawingState) HandleMessage(g *Game, m *Message) state {
	switch m.Action {
	case quit:
		state, err := g.handleQuit(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	case kick:
		state, err := g.handleKick(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	case chat:
		err := g.handleChat(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return nil
	case draw, erase, lineDraw, rectDraw, rectFill, circleDraw, circleFill,
		changeColor, changePencilSize, changeEraserSize, clearBoard:
		err := g.handleBoardAction(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return nil
	case guess:
		state, err := g.handleGuess(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	default:
		return nil
	}
}

type endingState struct {
	timer *time.Timer
}

func (s *endingState) Enter(g *Game) {
	slog.Info("Entering ending state")

	msg := newEmptyMessage(ending)
	g.sendToAll(msg) // FIXME

	s.timer = time.AfterFunc(15*time.Second, func() {
		g.stateCh <- &waitingState{}
	})
}

func (s *endingState) Exit(g *Game) {
	s.timer.Stop()
	slog.Info("Exiting ending state")
}

func (s *endingState) HandleMessage(g *Game, m *Message) state {
	switch m.Action {
	case quit:
		state, err := g.handleQuit(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	case kick:
		state, err := g.handleKick(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return state
	case chat:
		err := g.handleChat(m)
		if err != nil {
			slog.Error(err.Error())
		}
		return nil
	default:
		return nil
	}
}

type closingState struct {
}

func (s *closingState) Enter(g *Game) {
	slog.Info("Entering closing state")
	close(g.ch)
}

func (s *closingState) Exit(g *Game) {
	slog.Info("Exiting closing state")
}

func (s *closingState) HandleMessage(g *Game, m *Message) state {
	return nil
}
