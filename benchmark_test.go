package ddgo

import "testing"

func BenchmarkParseFirefox(b *testing.B) {
	b.ReportAllocs()
	detector, err := New(WithResultCacheSize(0))
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0"
	for i := 0; i < b.N; i++ {
		if _, err := detector.Parse(ua); err != nil {
			b.Fatalf("Parse() failed: %v", err)
		}
	}
}

func BenchmarkParseGooglebot(b *testing.B) {
	b.ReportAllocs()
	detector, err := New(WithResultCacheSize(0))
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}
	ua := "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	for i := 0; i < b.N; i++ {
		if _, err := detector.Parse(ua); err != nil {
			b.Fatalf("Parse() failed: %v", err)
		}
	}
}

func BenchmarkParseCachedFirefox(b *testing.B) {
	b.ReportAllocs()
	detector, err := New(WithResultCacheSize(256))
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0"
	for i := 0; i < b.N; i++ {
		if _, err := detector.Parse(ua); err != nil {
			b.Fatalf("Parse() failed: %v", err)
		}
	}
}

func BenchmarkParseWithHeaders(b *testing.B) {
	b.ReportAllocs()
	detector, err := New()
	if err != nil {
		b.Fatalf("New() failed: %v", err)
	}
	headers := map[string]string{
		"Sec-CH-UA-Full-Version-List": "\"Not A;Brand\";v=\"24\", \"Chromium\";v=\"122.0.6261.128\", \"Google Chrome\";v=\"122.0.6261.128\"",
		"Sec-CH-UA-Platform":          "\"Android\"",
		"Sec-CH-UA-Platform-Version":  "\"14.0.0\"",
		"Sec-CH-UA-Model":             "\"SM-G991B\"",
		"Sec-CH-UA-Mobile":            "?1",
	}
	for i := 0; i < b.N; i++ {
		if _, err := detector.ParseWithHeaders("Mozilla/5.0", headers); err != nil {
			b.Fatalf("ParseWithHeaders() failed: %v", err)
		}
	}
}
