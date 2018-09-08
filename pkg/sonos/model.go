package sonos

type refreshToken struct {
	AccessToken string `json:"access_token"`
}

// Household describe household notion
type Household struct {
	ID   string
	Name string
}
