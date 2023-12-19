package game

type action int

type Message struct {
	ID      string `json:"-"`
	Action  action `json:"action"`
	Payload string `json:"payload"`
}
