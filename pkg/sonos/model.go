package sonos

const (
	// Source constant for worker message
	Source = `sonos`
)

// Household is a set of players on the same network under an account
type Household struct {
	ID      string
	Name    string
	Groups  []*Group
	Players []*Player
}

// Player is a connected device
type Player struct {
	ID   string
	Name string
}

// Group is a set of players playing the same audio
type Group struct {
	ID            string
	Name          string
	PlaybackState string
	PlayerIds     []string
	Volume        *GroupVolume
}

// GroupVolume is the volume of a group
type GroupVolume struct {
	Volume int
	Muted  bool
	Fixed  bool
}

// Token describes refresh token response
type Token struct {
	AccessToken string `json:"access_token"`
}
