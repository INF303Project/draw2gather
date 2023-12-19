package game

import "sync"

var Hub GameHub

func init() {
	Hub = GameHub{
		games: make(map[string]*Game),
	}
}

type GameHub struct {
	games map[string]*Game
	mu    sync.RWMutex
}

func (h *GameHub) Get(id string) *Game {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.games[id]
}

func (h *GameHub) Set(id string, game *Game) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.games[id] = game
}

func (h *GameHub) Delete(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.games, id)
}
