package game

import "github.com/gorilla/websocket"

type Player struct {
	id   string
	conn *websocket.Conn

	game *Game
	ch   chan *Message
}

func NewPlayer(id string, conn *websocket.Conn, game *Game) *Player {
	return &Player{
		id:   id,
		conn: conn,
		ch:   make(chan *Message),
		game: game,
	}
}

func (p *Player) ReadPump() {
	for {
		var msg Message
		err := p.conn.ReadJSON(&msg)
		if err != nil {
			// Connection closed
			p.game.unregister <- p
			return
		}

		msg.ID = p.id
		p.game.ch <- &msg
	}
}

func (p *Player) WritePump() {
	for msg := range p.ch {
		err := p.conn.WriteJSON(msg)
		if err != nil {
			return
		}
	}
	// Game closed channel
	p.conn.WriteMessage(websocket.CloseMessage, []byte{})
	p.conn.Close()
}
