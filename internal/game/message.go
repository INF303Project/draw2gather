package game

type action int

const (
	draw action = iota
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

	start
	quit
	pick
	guess
	chat
	kick
)

type Message struct {
	player  *Player `json:"-"`
	Action  action  `json:"action"`
	Payload string  `json:"payload"`
}

func (m *Message) decodeMessage() (any, error) {
	return nil, nil
}
