package dyson

// Device for Dyson Link
type Device struct {
	LocalCredentials string
	Name             string
	ProductType      string
	ScaleUnit        string
	Serial           string
	Credentials      *Credentials
}

// Credentials contains device's credential
type Credentials struct {
	Serial       string `json:"serial"`
	PasswordHash string `json:"apPasswordHash"`
}

// Data stores data fo hub
type Data struct {
	Devices []*Device
}
