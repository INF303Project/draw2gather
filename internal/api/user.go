package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
)

func (h *apiHandler) handleUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createUser(w, r)
	case http.MethodGet:
		h.getUser(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

type createUserReq struct {
	Name string `json:"name"`
}

type createUserResp struct {
	Name string `json:"name"`
}

// POST /user
func (h *apiHandler) createUser(w http.ResponseWriter, r *http.Request) {
	if h.sessions.GetString(r.Context(), "game_id") != "" {
		http.Error(w, "already in game", http.StatusUnauthorized)
		return
	}

	var req createUserReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if h.sessions.GetString(r.Context(), "player_id") == "" {
		h.sessions.Put(r.Context(), "player_id", uuid.NewString())
	}
	h.sessions.Put(r.Context(), "name", req.Name)
	h.sessions.RenewToken(r.Context())

	err = json.NewEncoder(w).Encode(&createUserResp{
		Name: req.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type getUserResp struct {
	Name   string `json:"name"`
	UserID string `json:"user_id,omitempty"`
	GameID string `json:"game_id,omitempty"`
}

// GET /user
func (h *apiHandler) getUser(w http.ResponseWriter, r *http.Request) {
	name := h.sessions.GetString(r.Context(), "name")
	userID := h.sessions.GetString(r.Context(), "user_id")
	gameID := h.sessions.GetString(r.Context(), "game_id")

	if name == "" {
		nameID := rand.Intn(1000)
		name = fmt.Sprintf("User%d", nameID)
		h.sessions.Put(r.Context(), "player_id", uuid.NewString())
		h.sessions.Put(r.Context(), "name", name)
	}

	err := json.NewEncoder(w).Encode(&getUserResp{
		Name:   name,
		UserID: userID,
		GameID: gameID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
