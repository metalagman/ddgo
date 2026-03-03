package ddgo

// Unknown is a sentinel value used when parser data is not available.
// Callers can compare string fields to this value instead of empty string.
const Unknown = "Unknown"

// Result is the aggregate detection output returned by Parse variants.
type Result struct {
	UserAgent string
	Bot       Bot
	Client    Client
	OS        OS
	Device    Device
}

// Producer stores metadata about a bot vendor.
type Producer struct {
	Name string
	URL  string
}

// Bot describes bot detection output.
type Bot struct {
	IsBot    bool
	Name     string
	Category string
	URL      string
	Producer Producer
}

// Client describes client application output.
type Client struct {
	Type          string
	Name          string
	Version       string
	Engine        string
	EngineVersion string
}

// OS describes operating system output.
type OS struct {
	Name     string
	Version  string
	Platform string
}

// Device describes detected device output.
type Device struct {
	Type  string
	Brand string
	Model string
}
