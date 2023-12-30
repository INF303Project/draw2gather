package game

import (
	"encoding/json"
)

type action int

const (
	greet action = iota

	// General actions
	join
	quit
	kick
	chat

	newOwner

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
	updateScore

	// State actions
	waiting
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

type payloadType interface {
	string | int | []int | [2]string | *Player |
		pointsPayload | messagePayload | scorePayload | gamePayload
}

type payload[T payloadType] struct {
	Value T `json:"value,omitempty"`
}

type point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type pointsPayload struct {
	Start point `json:"start"`
	End   point `json:"end"`
}

type messagePayload struct {
	Player  string `json:"player"`
	Message string `json:"message"`
}

type scorePayload struct {
	Player string `json:"player"`
	Score  int    `json:"score"`
}

type gamePayload struct {
	State         string             `json:"state"`
	Player        string             `json:"player"`
	Owner         string             `json:"owner"`
	CurrentPlayer string             `json:"current_player"`
	Players       map[string]*Player `json:"players"`
	Commands      []*Message         `json:"commands"`
}

func newEmptyMessage(act action) *Message {
	return &Message{
		Action: act,
	}
}

func newMessage[T payloadType](act action, val T) *Message {
	payload, _ := json.Marshal(&payload[T]{
		Value: val,
	})
	return &Message{
		Action:  act,
		Payload: string(payload),
	}
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

	default:
		return nil, errInvalidAction
	}
}
