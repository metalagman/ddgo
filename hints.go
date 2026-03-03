package ddgo

import (
	"regexp"
	"strings"
)

var reClientHintBrand = regexp.MustCompile(`"([^"]+)"\s*;\s*v="([^"]+)"`)

const (
	clientHintBrandMatchGroups = 3
	priorityEdge               = 100
	priorityOpera              = 90
	priorityChrome             = 80
	priorityChromium           = 70
	priorityFirefox            = 60
	prioritySafari             = 50
)

// ClientHintBrand represents one structured brand entry from Sec-CH-UA.
type ClientHintBrand struct {
	Name    string
	Version string
}

// ClientHints contains normalized Sec-CH-UA style client hints.
//
// Mobile is nil when the client did not send Sec-CH-UA-Mobile.
type ClientHints struct {
	Brands          []ClientHintBrand
	Platform        string
	PlatformVersion string
	Model           string
	Mobile          *bool
}

// ParseClientHintsFromHeaders extracts known client hints from HTTP header
// values. Header name matching is case-insensitive.
func ParseClientHintsFromHeaders(headers map[string]string) ClientHints {
	if len(headers) == 0 {
		return ClientHints{}
	}

	brandHeader := headerValue(headers, "Sec-CH-UA-Full-Version-List")
	if brandHeader == "" {
		brandHeader = headerValue(headers, "Sec-CH-UA")
	}

	hints := ClientHints{
		Brands:          parseClientHintBrands(brandHeader),
		Platform:        normalizeClientHintToken(headerValue(headers, "Sec-CH-UA-Platform")),
		PlatformVersion: normalizeClientHintToken(headerValue(headers, "Sec-CH-UA-Platform-Version")),
		Model:           normalizeClientHintToken(headerValue(headers, "Sec-CH-UA-Model")),
	}
	if mobile, ok := parseClientHintMobile(headerValue(headers, "Sec-CH-UA-Mobile")); ok {
		hints.Mobile = &mobile
	}
	return hints
}

func applyClientHints(result *Result, hints ClientHints) {
	applyClientHintClient(&result.Client, hints.Brands)
	applyClientHintOS(&result.OS, hints)
	applyClientHintDevice(&result.Device, result.OS, hints)
}

func applyClientHintClient(client *Client, brands []ClientHintBrand) {
	brand, ok := pickClientBrand(brands)
	if !ok {
		return
	}
	profile, ok := profileForBrand(brand.Name)
	if !ok {
		return
	}

	if client.Name == "" {
		client.Type = profile.ClientType
		client.Name = profile.Name
		client.Engine = profile.Engine
	}
	if profile.Name == canonicalClientName(client.Name) && client.Version == "" {
		client.Version = brand.Version
	}
	if profile.Name == canonicalClientName(client.Name) && client.EngineVersion == "" {
		client.EngineVersion = brand.Version
	}
}

func applyClientHintOS(os *OS, hints ClientHints) {
	if os.Name == OSNameUnknown && hints.Platform != "" {
		os.Name = canonicalOSName(hints.Platform)
	}
	if os.Version == "" && hints.PlatformVersion != "" {
		os.Version = normalizeVersion(hints.PlatformVersion)
	}
	if os.Platform == PlatformUnknown {
		os.Platform = platformForOS(os.Name)
	}
}

func applyClientHintDevice(device *Device, os OS, hints ClientHints) {
	if device.Type == DeviceTypeUnknown {
		device.Type = inferDeviceType(hints.Mobile, os.Name)
	}
	if device.Model == "" && hints.Model != "" {
		device.Model = hints.Model
	}
	if device.Brand == "" {
		device.Brand = inferBrandFromModel(device.Model)
	}
}

func parseClientHintBrands(raw string) []ClientHintBrand {
	matches := reClientHintBrand.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return nil
	}
	brands := make([]ClientHintBrand, 0, len(matches))
	for _, match := range matches {
		if len(match) < clientHintBrandMatchGroups {
			continue
		}
		brand := normalizeClientHintToken(match[1])
		version := normalizeClientHintToken(match[2])
		if brand == "" || version == "" {
			continue
		}
		brands = append(brands, ClientHintBrand{
			Name:    brand,
			Version: version,
		})
	}
	return brands
}

func parseClientHintMobile(raw string) (mobile bool, ok bool) {
	switch strings.TrimSpace(raw) {
	case "?1", "\"?1\"", "1", "true":
		return true, true
	case "?0", "\"?0\"", "0", "false":
		return false, true
	default:
		return false, false
	}
}

func normalizeClientHintToken(raw string) string {
	return strings.Trim(strings.TrimSpace(raw), `"`)
}

func headerValue(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

type clientProfile struct {
	Name       string
	ClientType ClientType
	Engine     string
	Priority   int
}

func pickClientBrand(brands []ClientHintBrand) (ClientHintBrand, bool) {
	best := ClientHintBrand{}
	bestPriority := -1
	for _, brand := range brands {
		profile, ok := profileForBrand(brand.Name)
		if !ok || profile.Priority < bestPriority {
			continue
		}
		best = brand
		bestPriority = profile.Priority
	}
	if bestPriority < 0 {
		return ClientHintBrand{}, false
	}
	return best, true
}

func profileForBrand(brand string) (clientProfile, bool) {
	switch strings.ToLower(strings.TrimSpace(brand)) {
	case "microsoft edge", "edge":
		return clientProfile{Name: "Microsoft Edge", ClientType: ClientTypeBrowser, Engine: "Blink", Priority: priorityEdge}, true
	case "opera":
		return clientProfile{Name: "Opera", ClientType: ClientTypeBrowser, Engine: "Blink", Priority: priorityOpera}, true
	case "google chrome", "chrome":
		return clientProfile{Name: "Chrome", ClientType: ClientTypeBrowser, Engine: "Blink", Priority: priorityChrome}, true
	case "chromium":
		return clientProfile{Name: "Chrome", ClientType: ClientTypeBrowser, Engine: "Blink", Priority: priorityChromium}, true
	case "mozilla firefox", "firefox":
		return clientProfile{Name: "Firefox", ClientType: ClientTypeBrowser, Engine: "Gecko", Priority: priorityFirefox}, true
	case "safari":
		return clientProfile{Name: "Safari", ClientType: ClientTypeBrowser, Engine: "WebKit", Priority: prioritySafari}, true
	default:
		return clientProfile{}, false
	}
}

func canonicalClientName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "microsoft edge", "edge":
		return "Microsoft Edge"
	case "opera":
		return "Opera"
	case "google chrome", "chrome", "chromium":
		return "Chrome"
	case "mozilla firefox", "firefox":
		return "Firefox"
	case "safari":
		return "Safari"
	default:
		return name
	}
}

func canonicalOSName(name string) OSName {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "windows":
		return OSNameWindows
	case "android":
		return OSNameAndroid
	case "ios":
		return OSNameIOS
	case "macos", "mac os", "mac os x":
		return OSNameMacOS
	case "linux":
		return OSNameLinux
	case "chrome os", "cros":
		return OSNameChromeOS
	default:
		return OSName(strings.TrimSpace(name))
	}
}

func platformForOS(osName OSName) Platform {
	switch osName {
	case OSNameAndroid, OSNameIOS:
		return PlatformARM
	case OSNameWindows, OSNameMacOS, OSNameLinux, OSNameChromeOS:
		return PlatformX64
	case OSNameUnknown:
		return PlatformUnknown
	}
	return PlatformUnknown
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return version
	}
	parts := strings.Split(version, ".")
	for len(parts) > 1 && parts[len(parts)-1] == "0" {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(parts, ".")
}

func inferDeviceType(mobile *bool, osName OSName) DeviceType {
	if mobile != nil {
		if *mobile {
			return DeviceTypeSmartphone
		}
		switch osName {
		case OSNameAndroid, OSNameIOS:
			return DeviceTypeTablet
		case OSNameWindows, OSNameMacOS, OSNameLinux, OSNameChromeOS:
			return DeviceTypeDesktop
		case OSNameUnknown:
			return DeviceTypeUnknown
		}
		return DeviceTypeUnknown
	}

	switch osName {
	case OSNameAndroid, OSNameIOS:
		return DeviceTypeSmartphone
	case OSNameWindows, OSNameMacOS, OSNameLinux, OSNameChromeOS:
		return DeviceTypeDesktop
	case OSNameUnknown:
		return DeviceTypeUnknown
	}
	return DeviceTypeUnknown
}

func inferBrandFromModel(model string) string {
	if model == "" {
		return ""
	}
	upper := strings.ToUpper(model)
	switch {
	case strings.HasPrefix(upper, "SM-"):
		return "Samsung"
	case strings.HasPrefix(strings.ToLower(model), "pixel"):
		return "Google"
	case strings.Contains(strings.ToLower(model), "iphone"), strings.Contains(strings.ToLower(model), "ipad"):
		return "Apple"
	default:
		return ""
	}
}
