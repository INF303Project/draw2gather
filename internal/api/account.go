package api

import (
	"net/http"
	"strings"
)

func (h *apiHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.login(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (h *apiHandler) login(w http.ResponseWriter, r *http.Request) {
	playerID := h.sessions.GetString(r.Context(), "player_id")
	if playerID == "" {
		http.Error(w, "player_id is required", http.StatusUnauthorized)
		return
	}

	gameID := h.sessions.GetString(r.Context(), "game_id")
	if gameID != "" {
		http.Error(w, "already in game", http.StatusUnauthorized)
		return
	}

	auth := r.Header.Get("Authorization")
	bearer := strings.TrimPrefix(auth, "Bearer ")
	if bearer == "" {
		http.Error(w, "authorization header is required", http.StatusUnauthorized)
		return
	}

	token, err := h.auth.VerifyIDToken(r.Context(), bearer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	h.db.Collection("players").Doc(playerID).Create(r.Context(), nil)

	h.sessions.Put(r.Context(), "user_id", token.UID)
	h.sessions.RenewToken(r.Context())
}

func (h *apiHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.logout(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

func (h *apiHandler) logout(w http.ResponseWriter, r *http.Request) {
	if h.sessions.GetString(r.Context(), "user_id") == "" {
		http.Error(w, "not logged in", http.StatusUnauthorized)
		return
	}

	if h.sessions.GetString(r.Context(), "game_id") != "" {
		http.Error(w, "in game", http.StatusUnauthorized)
		return
	}

	h.sessions.Destroy(r.Context())
}
