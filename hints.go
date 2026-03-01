package ddgo

import (
	"regexp"
	"strings"
)

var reClientHintBrand = regexp.MustCompile(`"([^"]+)"\s*;\s*v="([^"]+)"`)

// ClientHintBrand represents one structured entry from Sec-CH-UA.
type ClientHintBrand struct {
	Name    string
	Version string
}

// ClientHints contains normalized Sec-CH-UA style client hints.
type ClientHints struct {
	Brands          []ClientHintBrand
	Platform        string
	PlatformVersion string
	Model           string
	Mobile          *bool
}

// ParseClientHintsFromHeaders extracts client hints from HTTP header values.
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

	if client.Name == Unknown {
		client.Type = profile.ClientType
		client.Name = profile.Name
		client.Engine = profile.Engine
	}
	if profile.Name == canonicalClientName(client.Name) && client.Version == Unknown {
		client.Version = brand.Version
	}
	if profile.Name == canonicalClientName(client.Name) && client.EngineVersion == Unknown {
		client.EngineVersion = brand.Version
	}
}

func applyClientHintOS(os *OS, hints ClientHints) {
	if os.Name == Unknown && hints.Platform != "" {
		os.Name = canonicalOSName(hints.Platform)
	}
	if os.Version == Unknown && hints.PlatformVersion != "" {
		os.Version = normalizeVersion(hints.PlatformVersion)
	}
	if os.Platform == Unknown {
		os.Platform = platformForOS(os.Name)
	}
}

func applyClientHintDevice(device *Device, os OS, hints ClientHints) {
	if device.Type == Unknown {
		device.Type = inferDeviceType(hints.Mobile, os.Name)
	}
	if device.Model == Unknown && hints.Model != "" {
		device.Model = hints.Model
	}
	if device.Brand == Unknown {
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
		if len(match) < 3 {
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

func parseClientHintMobile(raw string) (bool, bool) {
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
	ClientType string
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
		return clientProfile{Name: "Microsoft Edge", ClientType: "Browser", Engine: "Blink", Priority: 100}, true
	case "opera":
		return clientProfile{Name: "Opera", ClientType: "Browser", Engine: "Blink", Priority: 90}, true
	case "google chrome", "chrome":
		return clientProfile{Name: "Chrome", ClientType: "Browser", Engine: "Blink", Priority: 80}, true
	case "chromium":
		return clientProfile{Name: "Chrome", ClientType: "Browser", Engine: "Blink", Priority: 70}, true
	case "mozilla firefox", "firefox":
		return clientProfile{Name: "Firefox", ClientType: "Browser", Engine: "Gecko", Priority: 60}, true
	case "safari":
		return clientProfile{Name: "Safari", ClientType: "Browser", Engine: "WebKit", Priority: 50}, true
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

func canonicalOSName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "windows":
		return "Windows"
	case "android":
		return "Android"
	case "ios":
		return "iOS"
	case "macos", "mac os", "mac os x":
		return "macOS"
	case "linux":
		return "Linux"
	case "chrome os", "cros":
		return "Chrome OS"
	default:
		return name
	}
}

func platformForOS(osName string) string {
	switch osName {
	case "Android", "iOS":
		return "ARM"
	case "Windows", "macOS", "Linux", "Chrome OS":
		return "x64"
	default:
		return Unknown
	}
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

func inferDeviceType(mobile *bool, osName string) string {
	if mobile != nil {
		if *mobile {
			return "Smartphone"
		}
		switch osName {
		case "Android", "iOS":
			return "Tablet"
		case "Windows", "macOS", "Linux", "Chrome OS":
			return "Desktop"
		default:
			return Unknown
		}
	}

	switch osName {
	case "Android", "iOS":
		return "Smartphone"
	case "Windows", "macOS", "Linux", "Chrome OS":
		return "Desktop"
	default:
		return Unknown
	}
}

func inferBrandFromModel(model string) string {
	if model == Unknown || model == "" {
		return Unknown
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
		return Unknown
	}
}
