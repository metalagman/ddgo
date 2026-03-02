package ddsync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
)

var (
	errUpstreamVersionRequired      = errors.New("upstream version is required")
	errUpstreamRepoRequiredWithVer  = errors.New("upstream repo is required when upstream version is set")
	errSnapshotDirRequired          = errors.New("snapshot dir is required")
	errOutputPathRequired           = errors.New("output path is required")
	errManifestPathRequired         = errors.New("manifest path is required")
	errSnapshotVerificationFailed   = errors.New("snapshot verification failed")
	errManifestValidationConfigHint = errors.New("manifest status validation failed")
)

const (
	directoryPerm = 0o750
	filePerm      = 0o600
)

// Config controls input/output locations for the sync pipeline.
type Config struct {
	UpstreamRepo    string
	UpstreamVersion string
	SnapshotDir     string
	OutputPath      string
	ManifestPath    string
}

// SourceFile describes metadata about one input snapshot file.
type SourceFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// CompiledFile is the deterministic artifact entry for a source file.
type CompiledFile struct {
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	Content string `json:"content"`
}

// CompiledSnapshot is the generated artifact consumed by parsers.
type CompiledSnapshot struct {
	Files []CompiledFile `json:"files"`
}

// Manifest tracks source metadata and the expected output hash.
type Manifest struct {
	UpstreamRepo    string       `json:"upstream_repo"`
	UpstreamVersion string       `json:"upstream_version"`
	SourceDir       string       `json:"source_dir"`
	SourceFiles     []SourceFile `json:"source_files"`
	OutputPath      string       `json:"output_path"`
	OutputSHA256    string       `json:"output_sha256"`
}

// StatusReport describes snapshot/artifact consistency.
type StatusReport struct {
	Clean  bool
	Issues []string
}

// Update compiles snapshot inputs into deterministic output + manifest.
func Update(cfg Config) (Manifest, error) {
	err := validateConfigForUpdate(cfg)
	if err != nil {
		return Manifest{}, err
	}

	sourceFiles, compiledFiles, err := readSnapshot(cfg.SnapshotDir)
	if err != nil {
		return Manifest{}, err
	}

	compiled := CompiledSnapshot{
		Files: compiledFiles,
	}
	compiledBytes, err := marshalStableJSON(compiled)
	if err != nil {
		return Manifest{}, fmt.Errorf("marshal compiled snapshot: %w", err)
	}
	err = writeFile(cfg.OutputPath, compiledBytes)
	if err != nil {
		return Manifest{}, err
	}

	manifest := Manifest{
		UpstreamRepo:    cfg.UpstreamRepo,
		UpstreamVersion: cfg.UpstreamVersion,
		SourceDir:       toSlash(filepath.Clean(cfg.SnapshotDir)),
		SourceFiles:     sourceFiles,
		OutputPath:      toSlash(filepath.Clean(cfg.OutputPath)),
		OutputSHA256:    digestBytes(compiledBytes),
	}
	manifestBytes, err := marshalStableJSON(manifest)
	if err != nil {
		return Manifest{}, fmt.Errorf("marshal manifest: %w", err)
	}
	err = writeFile(cfg.ManifestPath, manifestBytes)
	if err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}

// Verify validates snapshot sources and output against the manifest.
func Verify(cfg Config) error {
	err := validateConfigForUpdate(cfg)
	if err != nil {
		return err
	}

	report, err := Status(cfg)
	if err != nil {
		return err
	}
	if report.Clean {
		return nil
	}
	return fmt.Errorf("%w: %s", errSnapshotVerificationFailed, strings.Join(report.Issues, "; "))
}

// Status checks whether current files match manifest expectations.
func Status(cfg Config) (StatusReport, error) {
	err := validateConfigForStatus(cfg)
	if err != nil {
		return StatusReport{}, err
	}

	manifestBytes, err := os.ReadFile(cfg.ManifestPath)
	if err != nil {
		return StatusReport{}, fmt.Errorf("read manifest: %w", err)
	}

	var manifest Manifest
	err = json.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		return StatusReport{}, fmt.Errorf("decode manifest: %w", err)
	}

	var issues []string
	if cfg.UpstreamRepo != "" {
		switch {
		case manifest.UpstreamRepo == "":
			issues = append(issues, "manifest missing upstream repo metadata")
		case manifest.UpstreamRepo != cfg.UpstreamRepo:
			issues = append(issues, fmt.Sprintf("upstream repo mismatch: want %q got %q", cfg.UpstreamRepo, manifest.UpstreamRepo))
		}
	}
	if cfg.UpstreamVersion != "" {
		switch {
		case manifest.UpstreamVersion == "":
			issues = append(issues, "manifest missing upstream version metadata")
		case manifest.UpstreamVersion != cfg.UpstreamVersion:
			issues = append(issues, fmt.Sprintf("upstream version mismatch: want %q got %q", cfg.UpstreamVersion, manifest.UpstreamVersion))
		}
	}

	sourceFiles, _, err := readSnapshot(cfg.SnapshotDir)
	if err != nil {
		return StatusReport{}, err
	}
	if !reflect.DeepEqual(sourceFiles, manifest.SourceFiles) {
		issues = append(issues, "source files differ from manifest")
	}

	compiledBytes, err := os.ReadFile(cfg.OutputPath)
	if err != nil {
		return StatusReport{}, fmt.Errorf("read output: %w", err)
	}
	if digestBytes(compiledBytes) != manifest.OutputSHA256 {
		issues = append(issues, "output artifact hash mismatch")
	}

	return StatusReport{
		Clean:  len(issues) == 0,
		Issues: issues,
	}, nil
}

func validateConfigForUpdate(cfg Config) error {
	err := validateConfigForStatus(cfg)
	if err != nil {
		return fmt.Errorf("%w: %w", errManifestValidationConfigHint, err)
	}
	if cfg.UpstreamVersion == "" {
		return errUpstreamVersionRequired
	}
	return nil
}

func validateConfigForStatus(cfg Config) error {
	if cfg.UpstreamVersion != "" && cfg.UpstreamRepo == "" {
		return errUpstreamRepoRequiredWithVer
	}
	if cfg.SnapshotDir == "" {
		return errSnapshotDirRequired
	}
	if cfg.OutputPath == "" {
		return errOutputPathRequired
	}
	if cfg.ManifestPath == "" {
		return errManifestPathRequired
	}
	return nil
}

func readSnapshot(snapshotDir string) ([]SourceFile, []CompiledFile, error) {
	var paths []string
	err := filepath.WalkDir(snapshotDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(snapshotDir, path)
		if err != nil {
			return err
		}
		paths = append(paths, toSlash(rel))
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walk snapshot: %w", err)
	}
	slices.Sort(paths)

	sourceFiles := make([]SourceFile, 0, len(paths))
	compiledFiles := make([]CompiledFile, 0, len(paths))
	for _, relPath := range paths {
		absPath := filepath.Join(snapshotDir, filepath.FromSlash(relPath))
		data, err := os.ReadFile(filepath.Clean(absPath))
		if err != nil {
			return nil, nil, fmt.Errorf("read %s: %w", relPath, err)
		}
		hash := digestBytes(data)
		sourceFiles = append(sourceFiles, SourceFile{
			Path:   relPath,
			SHA256: hash,
			Size:   int64(len(data)),
		})
		compiledFiles = append(compiledFiles, CompiledFile{
			Path:    relPath,
			SHA256:  hash,
			Content: string(data),
		})
	}

	return sourceFiles, compiledFiles, nil
}

func writeFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, directoryPerm)
	if err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	err = os.WriteFile(path, data, filePerm)
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func marshalStableJSON(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func toSlash(s string) string {
	return filepath.ToSlash(s)
}
