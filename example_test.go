package ddgo_test

import (
	"fmt"

	"github.com/metalagman/ddgo"
)

func ExampleNew() {
	detector := ddgo.New()
	result := detector.Parse("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0")

	fmt.Printf("%s %s\n", result.Client.Name, result.Client.Version)
	// Output:
	// Firefox 124.0
}

func ExampleDetector_Parse() {
	detector := ddgo.New()
	result := detector.Parse("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

	fmt.Printf("is_bot=%t name=%s device=%s\n", result.Bot.IsBot, result.Bot.Name, result.Device.Type)
	// Output:
	// is_bot=true name=Googlebot device=Bot
}

func ExampleDetector_ParseWithClientHints() {
	detector := ddgo.New()
	mobile := true
	hints := ddgo.ClientHints{
		Brands:          []ddgo.ClientHintBrand{{Name: "Google Chrome", Version: "122.0.6261.128"}},
		Platform:        "Android",
		PlatformVersion: "14.0.0",
		Model:           "SM-G991B",
		Mobile:          &mobile,
	}

	result := detector.ParseWithClientHints("Mozilla/5.0", hints)
	fmt.Printf("%s %s on %s (%s)\n", result.Client.Name, result.Client.Version, result.OS.Name, result.Device.Model)
	// Output:
	// Chrome 122.0.6261.128 on Android (SM-G991B)
}

func ExampleDetector_ParseWithHeaders() {
	detector := ddgo.New()
	headers := map[string]string{
		"Sec-CH-UA":                  "\"Not(A:Brand\";v=\"99\", \"Microsoft Edge\";v=\"123.0.0.0\", \"Chromium\";v=\"123.0.0.0\"",
		"Sec-CH-UA-Platform":         "\"Windows\"",
		"Sec-CH-UA-Platform-Version": "\"15.0.0\"",
		"Sec-CH-UA-Mobile":           "?0",
	}

	result := detector.ParseWithHeaders("Mozilla/5.0", headers)
	fmt.Printf("%s %s (%s)\n", result.Client.Name, result.Client.Version, result.Device.Type)
	// Output:
	// Microsoft Edge 123.0.0.0 (Desktop)
}

func ExampleWithMaxUserAgentLen() {
	detector := ddgo.New(ddgo.WithMaxUserAgentLen(7))
	result := detector.Parse("Mozilla/5.0")

	fmt.Println(result.UserAgent)
	// Output:
	// Mozilla
}

func ExampleWithUserAgentTrimming() {
	detector := ddgo.New(ddgo.WithUserAgentTrimming(false))
	result := detector.Parse("  Mozilla/5.0  ")

	fmt.Printf("%q\n", result.UserAgent)
	// Output:
	// "  Mozilla/5.0  "
}

func ExampleParseClientHintsFromHeaders() {
	headers := map[string]string{
		"Sec-CH-UA-Full-Version-List": "\"Not A;Brand\";v=\"24\", \"Chromium\";v=\"122.0.6261.128\", \"Google Chrome\";v=\"122.0.6261.128\"",
		"Sec-CH-UA-Platform":          "\"Android\"",
		"Sec-CH-UA-Mobile":            "?1",
	}

	hints := ddgo.ParseClientHintsFromHeaders(headers)
	fmt.Printf("brands=%d platform=%s mobile=%t\n", len(hints.Brands), hints.Platform, *hints.Mobile)
	// Output:
	// brands=3 platform=Android mobile=true
}
