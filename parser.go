package ddgo

import (
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

func parseBot(ua string) Bot {
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

func parseClient(ua string, isBot bool) Client {
	if isBot {
		return unknownClient()
	}

	if matches := reClientEdge.FindStringSubmatch(ua); len(matches) > 1 {
		return Client{
			Type:          "Browser",
			Name:          "Microsoft Edge",
			Version:       matches[1],
			Engine:        "Blink",
			EngineVersion: matches[1],
		}
	}
	if matches := reClientEdgeAlt.FindStringSubmatch(ua); len(matches) > 1 {
		engine := "Blink"
		engineVersion := matches[1]
		if strings.Contains(ua, "EdgiOS/") {
			engine = "WebKit"
			engineVersion = firstMatch(reWebKit, ua, matches[1])
		}
		return Client{
			Type:          "Browser",
			Name:          "Microsoft Edge",
			Version:       matches[1],
			Engine:        engine,
			EngineVersion: engineVersion,
		}
	}
	if matches := reClientOpera.FindStringSubmatch(ua); len(matches) > 1 {
		return Client{
			Type:          "Browser",
			Name:          "Opera",
			Version:       matches[1],
			Engine:        "Blink",
			EngineVersion: matches[1],
		}
	}
	if matches := reClientFirefoxOS.FindStringSubmatch(ua); len(matches) > 1 {
		return Client{
			Type:          "Browser",
			Name:          "Firefox",
			Version:       matches[1],
			Engine:        "WebKit",
			EngineVersion: firstMatch(reWebKit, ua, matches[1]),
		}
	}
	if matches := reClientFirefox.FindStringSubmatch(ua); len(matches) > 1 {
		return Client{
			Type:          "Browser",
			Name:          "Firefox",
			Version:       matches[1],
			Engine:        "Gecko",
			EngineVersion: firstMatch(reGeckoRV, ua, matches[1]),
		}
	}
	if matches := reClientChromeOS.FindStringSubmatch(ua); len(matches) > 1 {
		return Client{
			Type:          "Browser",
			Name:          "Chrome",
			Version:       matches[1],
			Engine:        "WebKit",
			EngineVersion: firstMatch(reWebKit, ua, matches[1]),
		}
	}
	if matches := reClientChrome.FindStringSubmatch(ua); len(matches) > 1 {
		return Client{
			Type:          "Browser",
			Name:          "Chrome",
			Version:       matches[1],
			Engine:        "Blink",
			EngineVersion: matches[1],
		}
	}
	if matches := reClientSafari.FindStringSubmatch(ua); len(matches) > 1 {
		return Client{
			Type:          "Browser",
			Name:          "Safari",
			Version:       matches[1],
			Engine:        "WebKit",
			EngineVersion: firstMatch(reWebKit, ua, matches[1]),
		}
	}
	if matches := reClientCurl.FindStringSubmatch(ua); len(matches) > 1 {
		return Client{
			Type:          "Library",
			Name:          "curl",
			Version:       matches[1],
			Engine:        Unknown,
			EngineVersion: Unknown,
		}
	}
	if matches := reClientGoHTTP.FindStringSubmatch(ua); len(matches) > 1 {
		return Client{
			Type:          "Library",
			Name:          "Go HTTP Client",
			Version:       matches[1],
			Engine:        Unknown,
			EngineVersion: Unknown,
		}
	}

	return unknownClient()
}

func parseOS(ua string) OS {
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

func parseDevice(ua string, isBot bool) Device {
	if isBot {
		return Device{
			Type:  "Bot",
			Brand: Unknown,
			Model: Unknown,
		}
	}

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
		} else if matches := rePixelModel.FindStringSubmatch(ua); len(matches) > 1 {
			brand = "Google"
			model = matches[1]
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
