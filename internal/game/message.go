package game

import "encoding/json"

type action int

const (
	// General actions
	join action = iota
	quit
	kick
	chat

	// Waiting state actions
	start

	// Picking state actions
	pick

	// Drawing state actions
	draw
	erase
	lineDraw
	rectDraw
	rectFill
	circleDraw
	circleFill
	changeColor
	changePencilSize
	changeEraserSize
	clearBoard
	guess

	correctGuess

	// State actions
	starting
	picking
	drawing
	ending
)

type Message struct {
	player  *Player `json:"-"`
	Action  action  `json:"action"`
	Payload string  `json:"payload,omitempty"`
}

type payload[T any] struct {
	Value T `json:"value,omitempty"`
}

type userPayload[T any] struct {
	ID    string `json:"id"`
	Value T      `json:"value,omitempty"`
}

type point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type pointsPayload struct {
	Start point `json:"start"`
	End   point `json:"end"`
}

func (m *Message) decodeMessage() (any, error) {
	switch m.Action {
	case quit, start, clearBoard:
		return nil, nil

	case kick, chat, pick, changeColor, guess:
		var p payload[string]
		err := json.Unmarshal([]byte(m.Payload), &p)
		if err != nil {
			return nil, err
		}
		return p, nil

	case changePencilSize, changeEraserSize:
		var p payload[int]
		err := json.Unmarshal([]byte(m.Payload), &p)
		if err != nil {
			return nil, err
		}
		return p, nil

	case draw, erase:
		var p payload[[]int]
		err := json.Unmarshal([]byte(m.Payload), &p)
		if err != nil {
			return nil, err
		}
		return p, nil

	case lineDraw, rectDraw, rectFill, circleDraw, circleFill:
		var p payload[pointsPayload]
		err := json.Unmarshal([]byte(m.Payload), &p)
		if err != nil {
			return nil, err
		}
		return p, nil
	}

	return nil, nil
}
