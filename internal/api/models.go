package api

type gameObject struct {
	ID             string         `firestore:"-" json:"id"`
	Owner          string         `firestore:"owner" json:"-"`
	Visibility     bool           `firestore:"visibility" json:"visibility"`
	Language       string         `firestore:"language" json:"language"`
	TargetScore    int            `firestore:"target_score" json:"target_score"`
	MaxPlayers     int            `firestore:"max_players" json:"max_players"`
	CurrentPlayers int            `firestore:"current_players" json:"current_players"`
	BannedPlayers  map[string]any `firestore:"banned_players" json:"-"`
}
