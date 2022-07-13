package v2

// Webhooks contains details of a webhook
type Webhooks struct {
	URL     string `json:"url"`
	Event   string `json:"event"`
	Enabled bool   `json:"enabled"`
}
