package ddgo

func parseClient(ua string, isBot bool) Client {
	if isBot {
		return unknownClient()
	}
	if client, ok := parseClientSnapshot(ua); ok {
		return client
	}
	return parseClientLegacy(ua, false)
}

func parseOS(ua string) OS {
	if os, ok := parseOSSnapshot(ua); ok {
		return os
	}
	return parseOSLegacy(ua)
}

func parseDevice(ua string, isBot bool) Device {
	if isBot {
		return Device{
			Type:  "Bot",
			Brand: Unknown,
			Model: Unknown,
		}
	}
	if device, ok := parseDeviceSnapshot(ua); ok {
		return device
	}
	return parseDeviceLegacy(ua, false)
}
