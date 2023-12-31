package game

import (
	"context"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`

	ctx  context.Context `json:"-"`
	conn *websocket.Conn `json:"-"`

	game *Game         `json:"-"`
	ch   chan *Message `json:"-"`
}

func NewPlayer(id, name string, ctx context.Context, conn *websocket.Conn, game *Game) *Player {
	return &Player{
		ID:    id,
		Name:  name,
		Score: 0,
		ctx:   ctx,
		conn:  conn,
		ch:    make(chan *Message, 256),
		game:  game,
	}
}

func (p *Player) ReadPump() {
	defer func() {
		if !p.game.Closed() {
			p.game.ch <- &Message{
				player: p,
				Action: quit,
			}
		}
		p.conn.Close()
	}()

	for {
		var msg Message
		err := p.conn.ReadJSON(&msg)
		if err != nil {
			// Connection closed
			return
		}

		msg.player = p
		p.game.ch <- &msg
	}
}

func (p *Player) WritePump() {
	defer func() {
		p.conn.WriteMessage(websocket.CloseMessage, []byte{})
		p.conn.Close()
	}()

	for msg := range p.ch {
		err := p.conn.WriteJSON(msg)
		if err != nil {
			return
		}
	}
}
