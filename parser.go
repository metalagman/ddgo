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

func parseBot(runtime *parserRuntime, ua string, uaRunes []rune) (Bot, error) {
	for _, rule := range runtime.botRules {
		matched, matchErr := matchRegexp2RunesBool(rule.pattern, uaRunes)
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
			URL:      Unknown,
			Producer: Producer{
				Name: Unknown,
				URL:  Unknown,
			},
		}
	default:
		return Bot{
			IsBot:    false,
			Name:     Unknown,
			Category: Unknown,
			URL:      Unknown,
			Producer: Producer{
				Name: Unknown,
				URL:  Unknown,
			},
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
		Type:          "Browser",
		Name:          name,
		Version:       version,
		Engine:        engine,
		EngineVersion: engineVersion,
	}
}

func libraryClient(name, version string) Client {
	return Client{
		Type:          "Library",
		Name:          name,
		Version:       version,
		Engine:        Unknown,
		EngineVersion: Unknown,
	}
}

func parseOSLegacy(ua string) OS {
	if matches := reOSWindows.FindStringSubmatch(ua); len(matches) > 1 {
		return OS{
			Name:     "Windows",
			Version:  mapWindowsVersion(matches[1]),
			Platform: windowsPlatform(ua),
		}
	}
	if matches := reOSAndroid.FindStringSubmatch(ua); len(matches) > 1 {
		return OS{
			Name:     "Android",
			Version:  matches[1],
			Platform: "ARM",
		}
	}
	if matches := reOSIOS.FindStringSubmatch(ua); len(matches) > 1 {
		return OS{
			Name:     "iOS",
			Version:  strings.ReplaceAll(matches[1], "_", "."),
			Platform: "ARM",
		}
	}
	if matches := reOSMac.FindStringSubmatch(ua); len(matches) > 1 {
		return OS{
			Name:     "macOS",
			Version:  strings.ReplaceAll(matches[1], "_", "."),
			Platform: "x64",
		}
	}
	if strings.Contains(ua, "Linux") {
		return OS{
			Name:     "Linux",
			Version:  Unknown,
			Platform: "x64",
		}
	}

	return OS{
		Name:     Unknown,
		Version:  Unknown,
		Platform: Unknown,
	}
}

func parseDeviceLegacy(ua string) Device {
	switch {
	case strings.Contains(ua, "iPad"):
		return Device{
			Type:  "Tablet",
			Brand: "Apple",
			Model: "iPad",
		}
	case strings.Contains(ua, "iPhone"):
		return Device{
			Type:  "Smartphone",
			Brand: "Apple",
			Model: "iPhone",
		}
	case strings.Contains(ua, "Android"):
		deviceType := "Tablet"
		if strings.Contains(ua, "Mobile") {
			deviceType = "Smartphone"
		}

		brand := Unknown
		model := Unknown
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
			Type:  "Desktop",
			Brand: Unknown,
			Model: Unknown,
		}
	default:
		return Device{
			Type:  Unknown,
			Brand: Unknown,
			Model: Unknown,
		}
	}
}

func unknownClient() Client {
	return Client{
		Type:          Unknown,
		Name:          Unknown,
		Version:       Unknown,
		Engine:        Unknown,
		EngineVersion: Unknown,
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

func windowsPlatform(ua string) string {
	switch {
	case strings.Contains(ua, "Win64"), strings.Contains(ua, "x64"), strings.Contains(ua, "WOW64"):
		return "x64"
	default:
		return "x86"
	}
}
