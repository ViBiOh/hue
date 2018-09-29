package sonos

type refreshToken struct {
	AccessToken string `json:"access_token"`
}

// Household is a set of players on the same network under an account
type Household struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Groups  []*Group  `json:"groups"`
	Players []*Player `json:"players"`
}

// Player is a connected device
type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Group is a set of players playing the same audio
type Group struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	PlaybackState string       `json:"playbackState"`
	PlayerIds     []string     `json:"playersIds"`
	Volume        *GroupVolume `json:"volume"`
}

// GroupVolume is the volume of a group
type GroupVolume struct {
	Volume int  `json:"value"`
	Muted  bool `json:"muted"`
	Fixed  bool `json:"fixed"`
}
