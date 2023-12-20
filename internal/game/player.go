package game

import (
	"context"

	"github.com/gorilla/websocket"
)

type Player struct {
	id   string
	name string

	ctx  context.Context
	conn *websocket.Conn

	game *Game
	ch   chan *Message
}

func NewPlayer(id, name string, ctx context.Context, conn *websocket.Conn, game *Game) *Player {
	return &Player{
		id:   id,
		name: name,
		ctx:  ctx,
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

		msg.player = p
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
