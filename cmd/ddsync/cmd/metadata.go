package ddcmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/metalagman/ddgo/internal/ddsync"
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

	payload, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal provenance: %w", err)
	}
	payload = append(payload, '\n')

	if err := os.MkdirAll(filepath.Dir(provenancePath), 0o755); err != nil {
		return fmt.Errorf("mkdir provenance dir: %w", err)
	}
	if err := os.WriteFile(provenancePath, payload, 0o644); err != nil {
		return fmt.Errorf("write provenance: %w", err)
	}
	return nil
}

func provenanceIssues(cfg ddsync.Config, provenancePath string) ([]string, error) {
	raw, err := os.ReadFile(provenancePath)
	if err != nil {
		return nil, fmt.Errorf("read provenance: %w", err)
	}

	var p provenance
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("decode provenance: %w", err)
	}

	issues := make([]string, 0, 6)

	var src *provenanceSource
	for i := range p.Sources {
		if p.Sources[i].Name == cfg.UpstreamRepo {
			src = &p.Sources[i]
			break
		}
	}
	if src == nil {
		issues = append(issues, fmt.Sprintf("provenance missing source entry for upstream repo %q", cfg.UpstreamRepo))
	} else {
		if src.UpstreamVersion == "" {
			issues = append(issues, "provenance missing upstream version metadata")
		} else if src.UpstreamVersion != cfg.UpstreamVersion {
			issues = append(issues, fmt.Sprintf("provenance upstream version mismatch: want %q got %q", cfg.UpstreamVersion, src.UpstreamVersion))
		}
		if src.SnapshotDir != normalizePath(cfg.SnapshotDir) {
			issues = append(issues, fmt.Sprintf("provenance snapshot dir mismatch: want %q got %q", normalizePath(cfg.SnapshotDir), src.SnapshotDir))
		}
	}

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

	var artifact *provenanceArtifact
	for i := range p.Artifacts {
		if p.Artifacts[i].Path == outputPath && p.Artifacts[i].ManifestPath == manifestPath {
			artifact = &p.Artifacts[i]
			break
		}
	}
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

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func normalizePath(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}
