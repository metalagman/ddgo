package ddgo

import (
	_ "embed"
	"encoding/json"
	"errors"
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

var errCompiledSnapshotEmptyPath = errors.New("compiled snapshot contains file with empty path")

func loadSnapshotFiles() (map[string]string, error) {
	var snapshot compiledSnapshot
	err := json.Unmarshal(compiledSnapshotJSON, &snapshot)
	if err != nil {
		return nil, err
	}

	files := make(map[string]string, len(snapshot.Files))
	for _, file := range snapshot.Files {
		if file.Path == "" {
			return nil, errCompiledSnapshotEmptyPath
		}
		files[file.Path] = file.Content
	}
	return files, nil
}
