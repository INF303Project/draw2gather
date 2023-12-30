package api

import (
	"encoding/json"
	"net/http"

	"cloud.google.com/go/firestore"
)

func (h *apiHandler) handleSet(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createWordSet(w, r)
	case http.MethodGet:
		h.getWordSets(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type createWordSetReq struct {
	Name     string   `json:"name"`
	Language string   `json:"language"`
	Words    []string `json:"words"`
}

func (h *apiHandler) createWordSet(w http.ResponseWriter, r *http.Request) {
	userID := h.sessions.GetString(r.Context(), "user_id")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req createWordSetReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.Name == "default" {
		http.Error(w, "default is a reserved name", http.StatusBadRequest)
		return
	}

	if len(req.Words) < 50 {
		http.Error(w, "at least 50 words are required", http.StatusBadRequest)
		return
	}

	_, err = h.db.Collection("players").Doc(userID).
		Collection("word_sets").Doc(req.Name).
		Set(r.Context(), &wordSetObject{
			Name:     req.Name,
			Language: req.Language,
			Words:    req.Words,
		})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type getWordSetsResp struct {
	Total    int             `json:"total"`
	WordSets []wordSetObject `json:"word_sets"`
}

func (h *apiHandler) getWordSets(w http.ResponseWriter, r *http.Request) {
	userID := h.sessions.GetString(r.Context(), "user_id")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var (
		docs       []*firestore.DocumentSnapshot
		err        error
		lang       = r.URL.Query().Get("lang")
		collection = h.db.Collection("players").Doc(userID).Collection("word_sets")
	)

	if lang == "" {
		docs, err = collection.Documents(r.Context()).GetAll()
	} else {
		docs, err = collection.Where("language", "==", lang).
			Documents(r.Context()).GetAll()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := getWordSetsResp{
		Total:    len(docs),
		WordSets: make([]wordSetObject, 0, len(docs)),
	}
	for _, doc := range docs {
		var ws wordSetObject
		err = doc.DataTo(&ws)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.WordSets = append(resp.WordSets, ws)
	}

	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
