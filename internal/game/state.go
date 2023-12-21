package game

import (
	"encoding/json"
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
	slog.Debug("Entering waiting state")
}

func (s *waitingState) Exit(g *Game) {
	slog.Debug("Exiting waiting state")
}

func (s *waitingState) HandleMessage(g *Game, m *Message) state {
	switch m.Action {
	case quit:
		state, _ := g.handleQuit(m)
		return state
	case kick:
		state, _ := g.handleKick(m)
		return state
	case chat:
		g.handleChat(m)
		return nil
	case start:
		state, _ := g.handleStart(m)
		return state
	default:
		return nil
	}
}

type startingState struct {
}

func (s *startingState) Enter(g *Game) {
	slog.Debug("Entering starting state")

	g.sendToAll(&Message{
		Action: starting,
	})

	time.AfterFunc(5*time.Second, func() {
		g.stateCh <- &pickingState{}
	})
}

func (s *startingState) Exit(g *Game) {
	slog.Debug("Exiting starting state")
}

func (s *startingState) HandleMessage(g *Game, m *Message) state {
	switch m.Action {
	case quit:
		state, _ := g.handleQuit(m)
		return state
	case kick:
		state, _ := g.handleKick(m)
		return state
	case chat:
		g.handleChat(m)
		return nil
	default:
		return nil
	}
}

type pickingState struct {
	timer *time.Timer
}

func (s *pickingState) Enter(g *Game) {
	slog.Debug("Entering picking state")

	g.currentDrawer = g.pickPlayer()
	words := g.pickWords()
	payload, _ := json.Marshal(&words)

	g.sendExceptPlayer(g.currentDrawer, &Message{
		Action: picking,
	})
	g.sendToPlayer(g.currentDrawer, &Message{
		Action:  pick,
		Payload: string(payload),
	})

	s.timer = time.AfterFunc(10*time.Second, func() {
		g.stateCh <- &startingState{}
	})
}

func (s *pickingState) Exit(g *Game) {
	s.timer.Stop()
	slog.Debug("Exiting picking state")
}

func (s *pickingState) HandleMessage(g *Game, m *Message) state {
	switch m.Action {
	case pick:
		state, _ := g.handlePick(m)
		return state
	}
	return nil
}

type drawingState struct {
	timer *time.Timer
}

func (s *drawingState) Enter(g *Game) {
	slog.Debug("Entering drawing state")

	s.timer = time.AfterFunc(1*time.Minute, func() {
		g.stateCh <- &startingState{}
	})
}

func (s *drawingState) Exit(g *Game) {
	s.timer.Stop()
	slog.Debug("Exiting drawing state")
}

func (s *drawingState) HandleMessage(g *Game, m *Message) state {
	switch m.Action {
	case guess:
		state, _ := g.handleGuess(m)
		return state
	}
	return nil
}

type endingState struct {
}

func (s *endingState) Enter(g *Game) {
	slog.Debug("Entering ending state")
}

func (s *endingState) Exit(g *Game) {
	slog.Debug("Exiting ending state")
}

func (s *endingState) HandleMessage(g *Game, m *Message) state {
	return nil
}
