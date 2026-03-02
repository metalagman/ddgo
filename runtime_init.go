package ddgo

import "fmt"

type parserRuntime struct {
	botRules      []botRule
	clientRules   []clientRuleSet
	clientEngines []clientEngineRule
	osRules       []osRule
	deviceRules   []deviceRule
}

func newParserRuntime() (*parserRuntime, error) {
	files, err := loadSnapshotFiles()
	if err != nil {
		return nil, fmt.Errorf("load snapshot files: %w", err)
	}

	botRules, err := loadBotRules(files)
	if err != nil {
		return nil, fmt.Errorf("load bot rules: %w", err)
	}
	clientRules, err := loadClientRules(files)
	if err != nil {
		return nil, fmt.Errorf("load client rules: %w", err)
	}
	clientEngineRules, err := loadClientEngineRules(files)
	if err != nil {
		return nil, fmt.Errorf("load client engine rules: %w", err)
	}
	osRules, err := loadOSRules(files)
	if err != nil {
		return nil, fmt.Errorf("load os rules: %w", err)
	}
	deviceRules, err := loadDeviceRules(files)
	if err != nil {
		return nil, fmt.Errorf("load device rules: %w", err)
	}

	return &parserRuntime{
		botRules:      botRules,
		clientRules:   clientRules,
		clientEngines: clientEngineRules,
		osRules:       osRules,
		deviceRules:   deviceRules,
	}, nil
}
