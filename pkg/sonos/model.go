package sonos

type refreshToken struct {
	AccessToken string `json:"access_token"`
}

// Household describe household notion
type Household struct {
	ID string
}

// Player describe player
type Player struct {
	ID   string
	Name string
}

// Group describe group notion
type Group struct {
	ID            string
	Name          string
	PlaybackState string
	PlayerIds     []string
}
