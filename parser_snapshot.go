package ddgo

func parseClient(runtime *parserRuntime, ua string) (Client, error) {
	client, ok, err := parseClientSnapshot(runtime, ua)
	if err != nil {
		return Client{}, err
	}
	if ok {
		return client, nil
	}
	return parseClientLegacy(ua), nil
}

func parseOS(runtime *parserRuntime, ua string) (OS, error) {
	osInfo, ok, err := parseOSSnapshot(runtime, ua)
	if err != nil {
		return OS{}, err
	}
	if ok {
		return osInfo, nil
	}
	return parseOSLegacy(ua), nil
}

func parseDevice(runtime *parserRuntime, ua string) (Device, error) {
	device, ok, err := parseDeviceSnapshot(runtime, ua)
	if err != nil {
		return Device{}, err
	}
	if ok {
		return device, nil
	}
	return parseDeviceLegacy(ua), nil
}
