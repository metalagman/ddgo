package ddgo

import "encoding/json"

// MarshalJSON keeps wire compatibility by rendering empty producer fields as "Unknown".
func (p Producer) MarshalJSON() ([]byte, error) {
	type wireProducer struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	wire := wireProducer{
		Name: marshalUnknownText(p.Name),
		URL:  marshalUnknownText(p.URL),
	}
	return json.Marshal(wire)
}

// UnmarshalJSON keeps wire compatibility by mapping "Unknown" producer fields to empty strings.
func (p *Producer) UnmarshalJSON(data []byte) error {
	type wireProducer struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	var wire wireProducer
	err := json.Unmarshal(data, &wire)
	if err != nil {
		return err
	}
	p.Name = unmarshalUnknownText(wire.Name)
	p.URL = unmarshalUnknownText(wire.URL)
	return nil
}

// MarshalJSON keeps wire compatibility by rendering empty bot text fields as "Unknown".
func (b Bot) MarshalJSON() ([]byte, error) {
	type wireBot struct {
		IsBot    bool     `json:"is_bot"`
		Name     string   `json:"name"`
		Category string   `json:"category"`
		URL      string   `json:"url"`
		Producer Producer `json:"producer"`
	}
	wire := wireBot{
		IsBot:    b.IsBot,
		Name:     marshalUnknownText(b.Name),
		Category: marshalUnknownText(b.Category),
		URL:      marshalUnknownText(b.URL),
		Producer: b.Producer,
	}
	return json.Marshal(wire)
}

// UnmarshalJSON keeps wire compatibility by mapping "Unknown" bot text fields to empty strings.
func (b *Bot) UnmarshalJSON(data []byte) error {
	type wireBot struct {
		IsBot    bool     `json:"is_bot"`
		Name     string   `json:"name"`
		Category string   `json:"category"`
		URL      string   `json:"url"`
		Producer Producer `json:"producer"`
	}

	var wire wireBot
	err := json.Unmarshal(data, &wire)
	if err != nil {
		return err
	}
	b.IsBot = wire.IsBot
	b.Name = unmarshalUnknownText(wire.Name)
	b.Category = unmarshalUnknownText(wire.Category)
	b.URL = unmarshalUnknownText(wire.URL)
	b.Producer = wire.Producer
	return nil
}

// MarshalJSON keeps wire compatibility by rendering empty client text fields as "Unknown".
func (c Client) MarshalJSON() ([]byte, error) {
	type wireClient struct {
		Type          ClientType `json:"type"`
		Name          string     `json:"name"`
		Version       string     `json:"version"`
		Engine        string     `json:"engine"`
		EngineVersion string     `json:"engine_version"`
	}
	wire := wireClient{
		Type:          normalizeClientTypeValue(c.Type),
		Name:          marshalUnknownText(c.Name),
		Version:       marshalUnknownText(c.Version),
		Engine:        marshalUnknownText(c.Engine),
		EngineVersion: marshalUnknownText(c.EngineVersion),
	}
	return json.Marshal(wire)
}

// UnmarshalJSON keeps wire compatibility by mapping "Unknown" client text fields to empty strings.
func (c *Client) UnmarshalJSON(data []byte) error {
	type wireClient struct {
		Type          ClientType `json:"type"`
		Name          string     `json:"name"`
		Version       string     `json:"version"`
		Engine        string     `json:"engine"`
		EngineVersion string     `json:"engine_version"`
	}

	var wire wireClient
	err := json.Unmarshal(data, &wire)
	if err != nil {
		return err
	}
	c.Type = normalizeClientTypeValue(wire.Type)
	c.Name = unmarshalUnknownText(wire.Name)
	c.Version = unmarshalUnknownText(wire.Version)
	c.Engine = unmarshalUnknownText(wire.Engine)
	c.EngineVersion = unmarshalUnknownText(wire.EngineVersion)
	return nil
}

// MarshalJSON keeps wire compatibility by rendering an empty OS version as "Unknown".
func (o OS) MarshalJSON() ([]byte, error) {
	type wireOS struct {
		Name     OSName   `json:"name"`
		Version  string   `json:"version"`
		Platform Platform `json:"platform"`
	}
	wire := wireOS{
		Name:     normalizeOSNameValue(o.Name),
		Version:  marshalUnknownText(o.Version),
		Platform: normalizePlatformValue(o.Platform),
	}
	return json.Marshal(wire)
}

// UnmarshalJSON keeps wire compatibility by mapping "Unknown" OS versions to empty strings.
func (o *OS) UnmarshalJSON(data []byte) error {
	type wireOS struct {
		Name     OSName   `json:"name"`
		Version  string   `json:"version"`
		Platform Platform `json:"platform"`
	}

	var wire wireOS
	err := json.Unmarshal(data, &wire)
	if err != nil {
		return err
	}
	o.Name = normalizeOSNameValue(wire.Name)
	o.Version = unmarshalUnknownText(wire.Version)
	o.Platform = normalizePlatformValue(wire.Platform)
	return nil
}

// MarshalJSON keeps wire compatibility by rendering empty device text fields as "Unknown".
func (d Device) MarshalJSON() ([]byte, error) {
	type wireDevice struct {
		Type  DeviceType `json:"type"`
		Brand string     `json:"brand"`
		Model string     `json:"model"`
	}
	wire := wireDevice{
		Type:  normalizeDeviceTypeValue(d.Type),
		Brand: marshalUnknownText(d.Brand),
		Model: marshalUnknownText(d.Model),
	}
	return json.Marshal(wire)
}

// UnmarshalJSON keeps wire compatibility by mapping "Unknown" device text fields to empty strings.
func (d *Device) UnmarshalJSON(data []byte) error {
	type wireDevice struct {
		Type  DeviceType `json:"type"`
		Brand string     `json:"brand"`
		Model string     `json:"model"`
	}

	var wire wireDevice
	err := json.Unmarshal(data, &wire)
	if err != nil {
		return err
	}
	d.Type = normalizeDeviceTypeValue(wire.Type)
	d.Brand = unmarshalUnknownText(wire.Brand)
	d.Model = unmarshalUnknownText(wire.Model)
	return nil
}
