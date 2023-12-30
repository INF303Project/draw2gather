package api

type gameObject struct {
	ID             string   `firestore:"-" json:"id"`
	Owner          string   `firestore:"owner" json:"-"`
	Visibility     bool     `firestore:"visibility" json:"visibility"`
	Language       string   `firestore:"language" json:"language"`
	TargetScore    int      `firestore:"target_score" json:"target_score"`
	MaxPlayers     int      `firestore:"max_players" json:"max_players"`
	CurrentPlayers []string `firestore:"current_players" json:"current_players"`
	BannedPlayers  []string `firestore:"banned_players" json:"-"`
}

type userObject struct {
}

type wordSetObject struct {
	Name     string   `firestore:"name" json:"name"`
	Language string   `firestore:"language" json:"language"`
	Words    []string `firestore:"words" json:"words"`
}
