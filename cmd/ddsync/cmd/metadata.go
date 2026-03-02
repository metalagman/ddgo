package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/metalagman/ddgo/internal/ddsync"
)

const (
	provenanceIssueCapacity = 6
	sourceIssueCapacity     = 2
	provenanceDirPerm       = 0o750
	provenanceFilePerm      = 0o600
)

type provenance struct {
	SchemaVersion int                  `json:"schema_version"`
	Sources       []provenanceSource   `json:"sources"`
	Artifacts     []provenanceArtifact `json:"artifacts"`
}

type provenanceSource struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	License         string `json:"license"`
	SnapshotDir     string `json:"snapshot_dir"`
	UpstreamVersion string `json:"upstream_version"`
}

type provenanceArtifact struct {
	Path         string `json:"path"`
	SHA256       string `json:"sha256"`
	ManifestPath string `json:"manifest_path"`
	ManifestSHA  string `json:"manifest_sha256"`
	Generator    string `json:"generator"`
}

func updateProvenance(cfg ddsync.Config, provenancePath string) error {
	outputSHA, err := fileSHA256(cfg.OutputPath)
	if err != nil {
		return fmt.Errorf("hash output artifact: %w", err)
	}
	manifestSHA, err := fileSHA256(cfg.ManifestPath)
	if err != nil {
		return fmt.Errorf("hash manifest: %w", err)
	}

	p := provenance{
		SchemaVersion: 1,
		Sources: []provenanceSource{
			{
				Name:            cfg.UpstreamRepo,
				URL:             "https://github.com/" + cfg.UpstreamRepo,
				License:         "LGPL-3.0-or-later",
				SnapshotDir:     normalizePath(cfg.SnapshotDir),
				UpstreamVersion: cfg.UpstreamVersion,
			},
		},
		Artifacts: []provenanceArtifact{
			{
				Path:         normalizePath(cfg.OutputPath),
				SHA256:       outputSHA,
				ManifestPath: normalizePath(cfg.ManifestPath),
				ManifestSHA:  manifestSHA,
				Generator: fmt.Sprintf(
					"go run ./cmd/ddsync update --upstream-repo %s",
					cfg.UpstreamRepo,
				),
			},
		},
	}

	payload, _ := json.MarshalIndent(p, "", "  ")
	payload = append(payload, '\n')

	err = os.MkdirAll(filepath.Dir(provenancePath), provenanceDirPerm)
	if err != nil {
		return fmt.Errorf("mkdir provenance dir: %w", err)
	}
	err = os.WriteFile(provenancePath, payload, provenanceFilePerm)
	if err != nil {
		return fmt.Errorf("write provenance: %w", err)
	}
	return nil
}

func provenanceIssues(cfg ddsync.Config, provenancePath string) ([]string, error) {
	raw, err := os.ReadFile(filepath.Clean(provenancePath))
	if err != nil {
		return nil, fmt.Errorf("read provenance: %w", err)
	}

	var p provenance
	err = json.Unmarshal(raw, &p)
	if err != nil {
		return nil, fmt.Errorf("decode provenance: %w", err)
	}

	issues := make([]string, 0, provenanceIssueCapacity)
	issues = append(issues, sourceIssues(p.Sources, cfg)...)

	outputPath := normalizePath(cfg.OutputPath)
	manifestPath := normalizePath(cfg.ManifestPath)
	outputSHA, err := fileSHA256(cfg.OutputPath)
	if err != nil {
		return nil, fmt.Errorf("hash output artifact: %w", err)
	}
	manifestSHA, err := fileSHA256(cfg.ManifestPath)
	if err != nil {
		return nil, fmt.Errorf("hash manifest: %w", err)
	}

	artifact := findProvenanceArtifact(p.Artifacts, outputPath, manifestPath)
	if artifact == nil {
		issues = append(issues, fmt.Sprintf("provenance missing artifact entry for %q", outputPath))
		return issues, nil
	}

	if artifact.SHA256 != outputSHA {
		issues = append(issues, "provenance output artifact hash mismatch")
	}
	if artifact.ManifestSHA != manifestSHA {
		issues = append(issues, "provenance manifest hash mismatch")
	}

	return issues, nil
}

func sourceIssues(sources []provenanceSource, cfg ddsync.Config) []string {
	src := findProvenanceSource(sources, cfg.UpstreamRepo)
	if src == nil {
		return []string{
			fmt.Sprintf("provenance missing source entry for upstream repo %q", cfg.UpstreamRepo),
		}
	}

	issues := make([]string, 0, sourceIssueCapacity)
	if src.UpstreamVersion == "" {
		issues = append(issues, "provenance missing upstream version metadata")
	} else if src.UpstreamVersion != cfg.UpstreamVersion {
		issues = append(
			issues,
			fmt.Sprintf(
				"provenance upstream version mismatch: want %q got %q",
				cfg.UpstreamVersion,
				src.UpstreamVersion,
			),
		)
	}
	expectedSnapshotDir := normalizePath(cfg.SnapshotDir)
	if src.SnapshotDir != expectedSnapshotDir {
		issues = append(
			issues,
			fmt.Sprintf(
				"provenance snapshot dir mismatch: want %q got %q",
				expectedSnapshotDir,
				src.SnapshotDir,
			),
		)
	}
	return issues
}

func findProvenanceSource(sources []provenanceSource, name string) *provenanceSource {
	for i := range sources {
		if sources[i].Name == name {
			return &sources[i]
		}
	}
	return nil
}

func findProvenanceArtifact(artifacts []provenanceArtifact, outputPath, manifestPath string) *provenanceArtifact {
	for i := range artifacts {
		if artifacts[i].Path == outputPath && artifacts[i].ManifestPath == manifestPath {
			return &artifacts[i]
		}
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func normalizePath(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}
