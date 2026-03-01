package ddgo

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed sync/compiled.json
var compiledSnapshotJSON []byte

type compiledSnapshotFile struct {
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	Content string `json:"content"`
}

type compiledSnapshot struct {
	Files []compiledSnapshotFile `json:"files"`
}

var (
	snapshotOnce  sync.Once
	snapshotFiles map[string]string
	snapshotErr   error
)

func loadSnapshotFiles() (map[string]string, error) {
	snapshotOnce.Do(func() {
		var snapshot compiledSnapshot
		if err := json.Unmarshal(compiledSnapshotJSON, &snapshot); err != nil {
			snapshotErr = fmt.Errorf("decode embedded compiled snapshot: %w", err)
			return
		}
		files := make(map[string]string, len(snapshot.Files))
		for _, file := range snapshot.Files {
			if file.Path == "" {
				snapshotErr = fmt.Errorf("compiled snapshot contains file with empty path")
				return
			}
			files[file.Path] = file.Content
		}
		snapshotFiles = files
	})

	if snapshotErr != nil {
		return nil, snapshotErr
	}
	return snapshotFiles, nil
}
