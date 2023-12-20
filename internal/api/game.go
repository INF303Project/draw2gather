package api

import (
	"context"
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
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

	for _, p := range g.BannedPlayers {
		if p == playerID {
			http.Error(w, "you are banned from this game", http.StatusForbidden)
			return
		}
	}

	_, err = h.db.Collection("games").Doc(req.GameID).Update(r.Context(), []firestore.Update{
		{
			Path: "current_players", Value: firestore.Increment(1),
		},
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
	name := h.sessions.GetString(r.Context(), "name")

	g := game.Hub.Get(gameID)
	if g == nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	ctx := context.WithoutCancel(r.Context())
	conn, err := h.ws.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	player := game.NewPlayer(playerID, name, ctx, conn, g)
	go player.ReadPump()
	go player.WritePump()
	g.Register(player)
}
