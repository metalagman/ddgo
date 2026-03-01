package ddgo

import "fmt"

func initParserRuntime() error {
	if _, err := loadSnapshotFiles(); err != nil {
		return fmt.Errorf("load snapshot files: %w", err)
	}
	if _, err := loadBotRules(); err != nil {
		return fmt.Errorf("load bot rules: %w", err)
	}
	if _, err := loadClientRules(); err != nil {
		return fmt.Errorf("load client rules: %w", err)
	}
	if _, err := loadClientEngineRules(); err != nil {
		return fmt.Errorf("load client engine rules: %w", err)
	}
	if _, err := loadOSRules(); err != nil {
		return fmt.Errorf("load os rules: %w", err)
	}
	if _, err := loadDeviceRules(); err != nil {
		return fmt.Errorf("load device rules: %w", err)
	}
	return nil
}
