package ddgo

func parseClient(ua string, isBot bool) (Client, error) {
	if isBot {
		return unknownClient(), nil
	}
	if client, ok, err := parseClientSnapshot(ua); err != nil {
		return Client{}, err
	} else if ok {
		return client, nil
	}
	return parseClientLegacy(ua, false), nil
}

func parseOS(ua string) (OS, error) {
	if os, ok, err := parseOSSnapshot(ua); err != nil {
		return OS{}, err
	} else if ok {
		return os, nil
	}
	return parseOSLegacy(ua), nil
}

func parseDevice(ua string, isBot bool) (Device, error) {
	if isBot {
		return Device{
			Type:  "Bot",
			Brand: Unknown,
			Model: Unknown,
		}, nil
	}
	if device, ok, err := parseDeviceSnapshot(ua); err != nil {
		return Device{}, err
	} else if ok {
		return device, nil
	}
	return parseDeviceLegacy(ua, false), nil
}
