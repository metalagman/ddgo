package ddgo

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reBotGoogle      = regexp.MustCompile(`(?i)\bGooglebot(?:/([0-9.]+))?\b`)
	reBotBing        = regexp.MustCompile(`(?i)\bbingbot(?:/([0-9.]+))?\b`)
	reBotDuckDuckBot = regexp.MustCompile(`(?i)\bDuckDuckBot(?:/([0-9.]+))?\b`)
	reBotAny         = regexp.MustCompile(`(?i)\b(bot|crawler|spider)\b`)

	reClientEdge      = regexp.MustCompile(`\bEdg/([0-9.]+)\b`)
	reClientEdgeAlt   = regexp.MustCompile(`\b(?:EdgA|EdgiOS)/([0-9.]+)\b`)
	reClientOpera     = regexp.MustCompile(`\bOPR/([0-9.]+)\b`)
	reClientFirefox   = regexp.MustCompile(`\bFirefox/([0-9.]+)\b`)
	reClientFirefoxOS = regexp.MustCompile(`\bFxiOS/([0-9.]+)\b`)
	reClientChrome    = regexp.MustCompile(`\bChrome/([0-9.]+)\b`)
	reClientChromeOS  = regexp.MustCompile(`\bCriOS/([0-9.]+)\b`)
	reClientSafari    = regexp.MustCompile(`\bVersion/([0-9.]+).*?\bSafari/`)
	reClientCurl      = regexp.MustCompile(`\bcurl/([0-9.]+)\b`)
	reClientGoHTTP    = regexp.MustCompile(`\bGo-http-client/([0-9.]+)\b`)
	reWebKit          = regexp.MustCompile(`\bAppleWebKit/([0-9.]+)\b`)
	reGeckoRV         = regexp.MustCompile(`\brv:([0-9.]+)\b`)

	reOSWindows = regexp.MustCompile(`\bWindows NT ([0-9.]+)\b`)
	reOSAndroid = regexp.MustCompile(`\bAndroid ([0-9.]+)\b`)
	reOSIOS     = regexp.MustCompile(`\b(?:CPU (?:iPhone )?OS|iPhone OS) ([0-9_]+)\b`)
	reOSMac     = regexp.MustCompile(`\bMac OS X ([0-9_]+)\b`)

	reSamsungModel = regexp.MustCompile(`\b(SM-[A-Z0-9]+)\b`)
	rePixelModel   = regexp.MustCompile(`\b(Pixel(?: [A-Za-z0-9]+)*)\b`)
)

func parseBot(runtime *parserRuntime, ua string) (Bot, error) {
	for _, rule := range runtime.botRules {
		_, matched, matchErr := matchRegexp2String(rule.pattern, ua)
		if matchErr != nil {
			return Bot{}, fmt.Errorf("match bot rule %q: %w", rule.name, matchErr)
		}
		if !matched {
			continue
		}
		return Bot{
			IsBot:    true,
			Name:     rule.name,
			Category: rule.category,
			URL:      rule.url,
			Producer: rule.producer,
		}, nil
	}

	return parseBotLegacy(ua), nil
}

func parseBotLegacy(ua string) Bot {
	switch {
	case reBotGoogle.MatchString(ua):
		return Bot{
			IsBot:    true,
			Name:     "Googlebot",
			Category: "Search bot",
			URL:      "https://www.google.com/bot.html",
			Producer: Producer{
				Name: "Google Inc.",
				URL:  "https://www.google.com",
			},
		}
	case reBotBing.MatchString(ua):
		return Bot{
			IsBot:    true,
			Name:     "bingbot",
			Category: "Search bot",
			URL:      "https://www.bing.com/webmasters/help/which-crawlers-does-bing-use-8c184ec0",
			Producer: Producer{
				Name: "Microsoft",
				URL:  "https://www.microsoft.com",
			},
		}
	case reBotDuckDuckBot.MatchString(ua):
		return Bot{
			IsBot:    true,
			Name:     "DuckDuckBot",
			Category: "Search bot",
			URL:      "https://duckduckgo.com/duckduckbot",
			Producer: Producer{
				Name: "DuckDuckGo",
				URL:  "https://duckduckgo.com",
			},
		}
	case reBotAny.MatchString(ua):
		return Bot{
			IsBot:    true,
			Name:     "Generic Bot",
			Category: "Bot",
			URL:      "",
			Producer: Producer{},
		}
	default:
		return Bot{
			IsBot:    false,
			Name:     "",
			Category: "",
			URL:      "",
			Producer: Producer{},
		}
	}
}

func parseClientLegacy(ua string) Client {
	client, ok := parseLegacyBrowserClient(ua)
	if ok {
		return client
	}
	client, ok = parseLegacyLibraryClient(ua)
	if ok {
		return client
	}
	return unknownClient()
}

func parseLegacyBrowserClient(ua string) (Client, bool) {
	matches := reClientEdge.FindStringSubmatch(ua)
	if len(matches) > 1 {
		return browserClient("Microsoft Edge", matches[1], "Blink", matches[1]), true
	}
	matches = reClientEdgeAlt.FindStringSubmatch(ua)
	if len(matches) > 1 {
		engine := "Blink"
		engineVersion := matches[1]
		if strings.Contains(ua, "EdgiOS/") {
			engine = "WebKit"
			engineVersion = firstMatch(reWebKit, ua, matches[1])
		}
		return browserClient("Microsoft Edge", matches[1], engine, engineVersion), true
	}
	matches = reClientOpera.FindStringSubmatch(ua)
	if len(matches) > 1 {
		return browserClient("Opera", matches[1], "Blink", matches[1]), true
	}
	matches = reClientFirefoxOS.FindStringSubmatch(ua)
	if len(matches) > 1 {
		return browserClient("Firefox", matches[1], "WebKit", firstMatch(reWebKit, ua, matches[1])), true
	}
	matches = reClientFirefox.FindStringSubmatch(ua)
	if len(matches) > 1 {
		return browserClient("Firefox", matches[1], "Gecko", firstMatch(reGeckoRV, ua, matches[1])), true
	}
	matches = reClientChromeOS.FindStringSubmatch(ua)
	if len(matches) > 1 {
		return browserClient("Chrome", matches[1], "WebKit", firstMatch(reWebKit, ua, matches[1])), true
	}
	matches = reClientChrome.FindStringSubmatch(ua)
	if len(matches) > 1 {
		return browserClient("Chrome", matches[1], "Blink", matches[1]), true
	}
	matches = reClientSafari.FindStringSubmatch(ua)
	if len(matches) > 1 {
		return browserClient("Safari", matches[1], "WebKit", firstMatch(reWebKit, ua, matches[1])), true
	}
	return Client{}, false
}

func parseLegacyLibraryClient(ua string) (Client, bool) {
	matches := reClientCurl.FindStringSubmatch(ua)
	if len(matches) > 1 {
		return libraryClient("curl", matches[1]), true
	}
	matches = reClientGoHTTP.FindStringSubmatch(ua)
	if len(matches) > 1 {
		return libraryClient("Go HTTP Client", matches[1]), true
	}
	return Client{}, false
}

func browserClient(name, version, engine, engineVersion string) Client {
	return Client{
		Type:          ClientTypeBrowser,
		Name:          name,
		Version:       version,
		Engine:        engine,
		EngineVersion: engineVersion,
	}
}

func libraryClient(name, version string) Client {
	return Client{
		Type:          ClientTypeLibrary,
		Name:          name,
		Version:       version,
		Engine:        "",
		EngineVersion: "",
	}
}

func parseOSLegacy(ua string) OS {
	if matches := reOSWindows.FindStringSubmatch(ua); len(matches) > 1 {
		return OS{
			Name:     OSNameWindows,
			Version:  mapWindowsVersion(matches[1]),
			Platform: windowsPlatform(ua),
		}
	}
	if matches := reOSAndroid.FindStringSubmatch(ua); len(matches) > 1 {
		return OS{
			Name:     OSNameAndroid,
			Version:  matches[1],
			Platform: PlatformARM,
		}
	}
	if matches := reOSIOS.FindStringSubmatch(ua); len(matches) > 1 {
		return OS{
			Name:     OSNameIOS,
			Version:  strings.ReplaceAll(matches[1], "_", "."),
			Platform: PlatformARM,
		}
	}
	if matches := reOSMac.FindStringSubmatch(ua); len(matches) > 1 {
		return OS{
			Name:     OSNameMacOS,
			Version:  strings.ReplaceAll(matches[1], "_", "."),
			Platform: PlatformX64,
		}
	}
	if strings.Contains(ua, "Linux") {
		return OS{
			Name:     OSNameLinux,
			Version:  "",
			Platform: PlatformX64,
		}
	}

	return OS{
		Name:     OSNameUnknown,
		Version:  "",
		Platform: PlatformUnknown,
	}
}

func parseDeviceLegacy(ua string) Device {
	switch {
	case strings.Contains(ua, "iPad"):
		return Device{
			Type:  DeviceTypeTablet,
			Brand: "Apple",
			Model: "iPad",
		}
	case strings.Contains(ua, "iPhone"):
		return Device{
			Type:  DeviceTypeSmartphone,
			Brand: "Apple",
			Model: "iPhone",
		}
	case strings.Contains(ua, "Android"):
		deviceType := DeviceTypeTablet
		if strings.Contains(ua, "Mobile") {
			deviceType = DeviceTypeSmartphone
		}

		brand := ""
		model := ""
		if matches := reSamsungModel.FindStringSubmatch(ua); len(matches) > 1 {
			brand = "Samsung"
			model = matches[1]
		} else if pixelMatches := rePixelModel.FindStringSubmatch(ua); len(pixelMatches) > 1 {
			brand = "Google"
			model = pixelMatches[1]
		}

		return Device{
			Type:  deviceType,
			Brand: brand,
			Model: model,
		}
	case strings.Contains(ua, "Windows NT"), strings.Contains(ua, "Macintosh"), strings.Contains(ua, "X11; Linux"):
		return Device{
			Type: DeviceTypeDesktop,
		}
	default:
		return Device{
			Type: DeviceTypeUnknown,
		}
	}
}

func unknownClient() Client {
	return Client{
		Type:          ClientTypeUnknown,
		Name:          "",
		Version:       "",
		Engine:        "",
		EngineVersion: "",
	}
}

func firstMatch(re *regexp.Regexp, s string, fallback string) string {
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return matches[1]
	}
	return fallback
}

func mapWindowsVersion(v string) string {
	switch v {
	case "10.0":
		return "10"
	case "6.3":
		return "8.1"
	case "6.2":
		return "8"
	case "6.1":
		return "7"
	default:
		return v
	}
}

func windowsPlatform(ua string) Platform {
	switch {
	case strings.Contains(ua, "Win64"), strings.Contains(ua, "x64"), strings.Contains(ua, "WOW64"):
		return PlatformX64
	default:
		return PlatformX86
	}
}
