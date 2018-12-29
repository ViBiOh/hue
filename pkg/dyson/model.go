package dyson

// Device for Dyson Link
type Device struct {
	LocalCredentials string
	Name             string
	ProductType      string
	ScaleUnit        string
	Serial           string
	Credentials      *Credentials
	State            *State
}

// Credentials contains device's credential
type Credentials struct {
	Serial       string `json:"serial"`
	PasswordHash string `json:"apPasswordHash"`
}

// State of device
type State struct {
	Temperature float32
	Humidity    float32
}

// Data stores data fo hub
type Data struct {
	Devices []*Device
}
