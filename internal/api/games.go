package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alperenunal/draw2gather/internal/game"
	"github.com/google/uuid"
)

func (h *apiHandler) handleGames(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createGame(w, r)
	case http.MethodGet:
		h.getGames(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

type createGameReq struct {
	Language    string `json:"language"`
	MaxPlayers  int    `json:"max_players"`
	TargetScore int    `json:"target_score"`
	Visibility  bool   `json:"visibility"`
}

type createGameResp struct {
	ID string `json:"id"`
}

// POST /games
func (h *apiHandler) createGame(w http.ResponseWriter, r *http.Request) {
	playerID := h.sessions.GetString(r.Context(), "player_id")
	if playerID == "" {
		http.Error(w, "player_id is required", http.StatusUnauthorized)
		return
	}

	gameID := h.sessions.GetString(r.Context(), "game_id")
	if gameID != "" {
		http.Error(w, "already in a game", http.StatusForbidden)
		return
	}

	var req createGameReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uuid.NewString()
	_, err = h.db.Collection("games").Doc(id).Set(r.Context(), &gameObject{
		Owner:          playerID,
		Visibility:     req.Visibility,
		Language:       req.Language,
		TargetScore:    req.TargetScore,
		MaxPlayers:     req.MaxPlayers,
		CurrentPlayers: 0,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(&createGameResp{
		ID: id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	g := game.NewGame(id, playerID, h.db)
	go g.Run()
	game.Hub.Set(id, g)
}

type getGamesResp struct {
	Total  int          `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
	Games  []gameObject `json:"games"`
}

// GET /games
func (h *apiHandler) getGames(w http.ResponseWriter, r *http.Request) {
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")
	if limitParam == "" {
		limitParam = "20"
	}
	if offsetParam == "" {
		offsetParam = "0"
	}

	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	offset, err := strconv.Atoi(offsetParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	language := r.URL.Query().Get("lang")

	query := h.db.Collection("games").Where("visibility", "==", true)
	if language != "" {
		query = query.Where("language", "==", language)
	}
	games, err := query.
		Limit(limit).
		Offset(offset).
		Documents(r.Context()).
		GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := getGamesResp{
		Total:  len(games),
		Limit:  limit,
		Offset: offset,
		Games:  make([]gameObject, 0, len(games)),
	}

	for _, doc := range games {
		var g gameObject
		err := doc.DataTo(&g)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		g.ID = doc.Ref.ID
		resp.Games = append(resp.Games, g)
	}

	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
