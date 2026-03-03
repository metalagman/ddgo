package ddgo

const unknownLabel = "Unknown"

// ClientType describes detected client category.
type ClientType string

const (
	// ClientTypeUnknown indicates that the client category could not be determined.
	ClientTypeUnknown ClientType = unknownLabel
	// ClientTypeBrowser indicates a web browser client.
	ClientTypeBrowser ClientType = "Browser"
	// ClientTypeLibrary indicates a software library client.
	ClientTypeLibrary ClientType = "Library"
	// ClientTypeFeedReader indicates a feed reader client.
	ClientTypeFeedReader ClientType = "Feed Reader"
	// ClientTypeMobileApp indicates a mobile application client.
	ClientTypeMobileApp ClientType = "Mobile App"
	// ClientTypeMediaPlayer indicates a media player client.
	ClientTypeMediaPlayer ClientType = "Media Player"
	// ClientTypePIM indicates a personal information manager client.
	ClientTypePIM ClientType = "PIM"
)

// DeviceType describes detected device category.
type DeviceType string

const (
	// DeviceTypeUnknown indicates that the device category could not be determined.
	DeviceTypeUnknown DeviceType = unknownLabel
	// DeviceTypeBot indicates an automated bot device.
	DeviceTypeBot DeviceType = "Bot"
	// DeviceTypeSmartphone indicates a smartphone device.
	DeviceTypeSmartphone DeviceType = "Smartphone"
	// DeviceTypeFeaturePhone indicates a feature phone device.
	DeviceTypeFeaturePhone DeviceType = "Feature Phone"
	// DeviceTypePhablet indicates a phablet device.
	DeviceTypePhablet DeviceType = "Phablet"
	// DeviceTypeTablet indicates a tablet device.
	DeviceTypeTablet DeviceType = "Tablet"
	// DeviceTypeDesktop indicates a desktop device.
	DeviceTypeDesktop DeviceType = "Desktop"
	// DeviceTypeConsole indicates a gaming console device.
	DeviceTypeConsole DeviceType = "Console"
	// DeviceTypeTV indicates a television device.
	DeviceTypeTV DeviceType = "TV"
	// DeviceTypeCamera indicates a camera device.
	DeviceTypeCamera DeviceType = "Camera"
	// DeviceTypeCarBrowser indicates an in-car browser device.
	DeviceTypeCarBrowser DeviceType = "Car Browser"
	// DeviceTypePortableMedia indicates a portable media player device.
	DeviceTypePortableMedia DeviceType = "Portable Media Player"
	// DeviceTypeSmartDisplay indicates a smart display device.
	DeviceTypeSmartDisplay DeviceType = "Smart Display"
	// DeviceTypeSmartSpeaker indicates a smart speaker device.
	DeviceTypeSmartSpeaker DeviceType = "Smart Speaker"
	// DeviceTypePeripheral indicates a peripheral device.
	DeviceTypePeripheral DeviceType = "Peripheral"
	// DeviceTypeWearable indicates a wearable device.
	DeviceTypeWearable DeviceType = "Wearable"
)

// OSName describes detected operating system name.
type OSName string

const (
	// OSNameUnknown indicates that the OS name could not be determined.
	OSNameUnknown OSName = unknownLabel
	// OSNameWindows indicates Microsoft Windows.
	OSNameWindows OSName = "Windows"
	// OSNameAndroid indicates Android.
	OSNameAndroid OSName = "Android"
	// OSNameIOS indicates Apple iOS.
	OSNameIOS OSName = "iOS"
	// OSNameMacOS indicates Apple macOS.
	OSNameMacOS OSName = "macOS"
	// OSNameLinux indicates Linux.
	OSNameLinux OSName = "Linux"
	// OSNameChromeOS indicates Google Chrome OS.
	OSNameChromeOS OSName = "Chrome OS"
)

// Platform describes detected architecture/platform.
type Platform string

const (
	// PlatformUnknown indicates that the platform could not be determined.
	PlatformUnknown Platform = unknownLabel
	// PlatformARM indicates ARM architecture.
	PlatformARM Platform = "ARM"
	// PlatformX64 indicates x86-64 architecture.
	PlatformX64 Platform = "x64"
	// PlatformX86 indicates x86 architecture.
	PlatformX86 Platform = "x86"
)

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
	Type          ClientType
	Name          string
	Version       string
	Engine        string
	EngineVersion string
}

// OS describes operating system output.
type OS struct {
	Name     OSName
	Version  string
	Platform Platform
}

// Device describes detected device output.
type Device struct {
	Type  DeviceType
	Brand string
	Model string
}

func normalizeClientTypeValue(value ClientType) ClientType {
	if value == "" {
		return ClientTypeUnknown
	}
	return value
}

func normalizeDeviceTypeValue(value DeviceType) DeviceType {
	if value == "" {
		return DeviceTypeUnknown
	}
	return value
}

func normalizeOSNameValue(value OSName) OSName {
	if value == "" {
		return OSNameUnknown
	}
	return value
}

func normalizePlatformValue(value Platform) Platform {
	if value == "" {
		return PlatformUnknown
	}
	return value
}

func marshalUnknownText(value string) string {
	if value == "" {
		return unknownLabel
	}
	return value
}

func unmarshalUnknownText(value string) string {
	if value == "" || value == unknownLabel {
		return ""
	}
	return value
}
