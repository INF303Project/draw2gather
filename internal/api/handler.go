package api

import (
	"context"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	scsfs "github.com/alexedwards/scs/firestore"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

type apiHandler struct {
	auth     *auth.Client
	db       *firestore.Client
	sessions *scs.SessionManager
	ws       websocket.Upgrader
}

func NewHandler(app *firebase.App) (http.Handler, error) {
	if err := os.MkdirAll("./logs/games", 0755); err != nil {
		return nil, err
	}

	ctx := context.Background()

	firestore, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	auth, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	store := scsfs.New(firestore)
	sessions := scs.New()
	sessions.Store = store
	sessions.Cookie.Name = "draw2gather"
	sessions.Cookie.SameSite = http.SameSiteStrictMode
	sessions.Cookie.HttpOnly = true
	sessions.Cookie.Secure = false

	ws := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	h := &apiHandler{
		auth:     auth,
		db:       firestore,
		sessions: sessions,
		ws:       ws,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/user", h.handleUser)
	mux.HandleFunc("/login", h.handleLogin)
	mux.HandleFunc("/logout", h.handleLogout)
	mux.HandleFunc("/set", h.handleSet)
	mux.HandleFunc("/games", h.handleGames)
	mux.HandleFunc("/game", h.handleGame)
	mux.HandleFunc("/health", h.handleHealth)

	handler := http.Handler(mux)
	handler = sessions.LoadAndSave(handler)
	handler = cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
	}).Handler(handler)

	return handler, nil
}

func (h *apiHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
