package api

import (
	"encoding/json"
	"net/http"

	"github.com/alperenunal/draw2gather/internal/game"
)

func (h *apiHandler) handleGame(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		h.joinGame(w, r)
	case http.MethodGet:
		h.joinServer(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

type joinGameReq struct {
	GameID string `json:"game_id"`
}

// PUT /game
func (h *apiHandler) joinGame(w http.ResponseWriter, r *http.Request) {
	var req joinGameReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	playerID := h.sessions.GetString(r.Context(), "player_id")
	if playerID == "" {
		http.Error(w, "player_id is required", http.StatusUnauthorized)
		return
	}

	gameID := h.sessions.GetString(r.Context(), "game_id")
	if gameID != "" {
		if gameID == req.GameID {
			return
		}
		http.Error(w, "already in a game", http.StatusForbidden)
		return
	}

	doc, err := h.db.Collection("games").Doc(req.GameID).Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var g gameObject
	err = doc.DataTo(&g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if g.CurrentPlayers >= g.MaxPlayers {
		http.Error(w, "game is full", http.StatusForbidden)
		return
	}

	if _, ok := g.BannedPlayers[playerID]; ok {
		http.Error(w, "you are banned from this game", http.StatusForbidden)
		return
	}

	_, err = h.db.Collection("games").Doc(req.GameID).Set(r.Context(), &gameObject{
		MaxPlayers:     g.MaxPlayers,
		TargetScore:    g.TargetScore,
		Language:       g.Language,
		Visibility:     g.Visibility,
		CurrentPlayers: g.CurrentPlayers + 1,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sessions.Put(r.Context(), "game_id", req.GameID)
}

// GET /game
func (h *apiHandler) joinServer(w http.ResponseWriter, r *http.Request) {
	gameID := h.sessions.GetString(r.Context(), "game_id")
	if gameID == "" {
		http.Error(w, "not in a game", http.StatusForbidden)
		return
	}
	playerID := h.sessions.GetString(r.Context(), "player_id")

	g := game.Hub.Get(gameID)
	if g == nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	conn, err := h.ws.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	player := game.NewPlayer(playerID, conn, g)
	go player.ReadPump()
	go player.WritePump()
	g.Register(player)
}
