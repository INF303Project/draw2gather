package api

import (
	"time"
)

type session struct {
	PlayerID string    `firestore:"player_id"`
	UserID   string    `firestore:"user_id,omitempty"`
	Name     string    `firestore:"name"`
	GameID   string    `firestore:"game_id,omitempty"`
	Expiry   time.Time `firestore:"expiry"`
}
