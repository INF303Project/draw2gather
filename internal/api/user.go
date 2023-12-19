package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

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

	auth := r.Header.Get("Authorization")
	bearer := strings.TrimPrefix(auth, "Bearer: ")

	if bearer != "" {
		token, err := h.auth.VerifyIDToken(r.Context(), bearer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		h.sessions.Put(r.Context(), "user_id", token.UID)
	}

	if h.sessions.GetString(r.Context(), "player_id") == "" {
		h.sessions.Put(r.Context(), "player_id", uuid.NewString())
		h.sessions.Put(r.Context(), "name", req.Name)
	} else {
		name := h.sessions.GetString(r.Context(), "name")
		if name != req.Name {
			h.sessions.Put(r.Context(), "name", req.Name)
		}
	}

	err = json.NewEncoder(w).Encode(&createUserResp{
		Name: req.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type getUserResp struct {
	UserID string `json:"user_id,omitempty"`
	Name   string `json:"name"`
}

// GET /user
func (h *apiHandler) getUser(w http.ResponseWriter, r *http.Request) {
	name := h.sessions.GetString(r.Context(), "name")
	userID := h.sessions.GetString(r.Context(), "user_id")

	if name == "" {
		nameID := rand.Intn(1000)
		name = fmt.Sprintf("User%d", nameID)
		h.sessions.Put(r.Context(), "player_id", uuid.NewString())
		h.sessions.Put(r.Context(), "name", name)
	}

	err := json.NewEncoder(w).Encode(&getUserResp{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
