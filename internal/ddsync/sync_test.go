package ddsync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateDeterministic(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	if _, err := Update(cfg); err != nil {
		t.Fatalf("Update() first run: %v", err)
	}

	outputBefore, err := os.ReadFile(cfg.OutputPath)
	if err != nil {
		t.Fatalf("read output before: %v", err)
	}
	manifestBefore, err := os.ReadFile(cfg.ManifestPath)
	if err != nil {
		t.Fatalf("read manifest before: %v", err)
	}

	if _, err := Update(cfg); err != nil {
		t.Fatalf("Update() second run: %v", err)
	}

	outputAfter, err := os.ReadFile(cfg.OutputPath)
	if err != nil {
		t.Fatalf("read output after: %v", err)
	}
	manifestAfter, err := os.ReadFile(cfg.ManifestPath)
	if err != nil {
		t.Fatalf("read manifest after: %v", err)
	}

	if string(outputBefore) != string(outputAfter) {
		t.Fatal("compiled artifact changed between identical update runs")
	}
	if string(manifestBefore) != string(manifestAfter) {
		t.Fatal("manifest changed between identical update runs")
	}
}

func TestVerifyDetectsDrift(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	if _, err := Update(cfg); err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
	if err := Verify(cfg); err != nil {
		t.Fatalf("Verify() clean case failed: %v", err)
	}

	sourceFile := filepath.Join(cfg.SnapshotDir, "bots.yml")
	if err := os.WriteFile(sourceFile, []byte("bot: altered\n"), 0o644); err != nil {
		t.Fatalf("mutate source: %v", err)
	}

	err := Verify(cfg)
	if err == nil {
		t.Fatal("Verify() expected drift error")
	}
	if !strings.Contains(err.Error(), "source files differ from manifest") {
		t.Fatalf("unexpected verify error: %v", err)
	}
}

func TestStatusDirtyWhenOutputTampered(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	if _, err := Update(cfg); err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	if err := os.WriteFile(cfg.OutputPath, []byte(`{"bad":"data"}`), 0o644); err != nil {
		t.Fatalf("tamper output: %v", err)
	}

	report, err := Status(cfg)
	if err != nil {
		t.Fatalf("Status() failed: %v", err)
	}
	if report.Clean {
		t.Fatal("Status() expected dirty state")
	}
	if len(report.Issues) == 0 {
		t.Fatal("Status() expected at least one issue")
	}
}

func TestVerifyRequiresUpstreamVersion(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	cfg.UpstreamVersion = ""
	err := Verify(cfg)
	if err == nil {
		t.Fatal("Verify() expected missing upstream version error")
	}
	if !strings.Contains(err.Error(), "upstream version is required") {
		t.Fatalf("unexpected verify error: %v", err)
	}
}

func TestStatusDetectsUpstreamVersionMismatch(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	if _, err := Update(cfg); err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	cfg.UpstreamVersion = "v-other"
	report, err := Status(cfg)
	if err != nil {
		t.Fatalf("Status() failed: %v", err)
	}
	if report.Clean {
		t.Fatalf("Status() expected dirty report: %+v", report)
	}

	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue, "upstream version mismatch") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected upstream version mismatch issue, got %+v", report.Issues)
	}
}

func TestStatusDetectsMissingUpstreamMetadata(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	if _, err := Update(cfg); err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	raw, err := os.ReadFile(cfg.ManifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var legacy map[string]any
	if err := json.Unmarshal(raw, &legacy); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	delete(legacy, "upstream_repo")
	delete(legacy, "upstream_version")
	legacyRaw, err := json.MarshalIndent(legacy, "", "  ")
	if err != nil {
		t.Fatalf("encode manifest: %v", err)
	}
	legacyRaw = append(legacyRaw, '\n')
	if err := os.WriteFile(cfg.ManifestPath, legacyRaw, 0o644); err != nil {
		t.Fatalf("write legacy manifest: %v", err)
	}

	report, err := Status(cfg)
	if err != nil {
		t.Fatalf("Status() failed: %v", err)
	}
	if report.Clean {
		t.Fatalf("expected dirty report, got %+v", report)
	}

	missingRepo := false
	missingVersion := false
	for _, issue := range report.Issues {
		if strings.Contains(issue, "manifest missing upstream repo metadata") {
			missingRepo = true
		}
		if strings.Contains(issue, "manifest missing upstream version metadata") {
			missingVersion = true
		}
	}
	if !missingRepo || !missingVersion {
		t.Fatalf("expected missing upstream metadata issues, got %+v", report.Issues)
	}
}

func testConfig(t *testing.T) Config {
	t.Helper()

	root := t.TempDir()
	snapshotDir := filepath.Join(root, "snapshots")
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatalf("mkdir snapshots: %v", err)
	}
	if err := os.WriteFile(filepath.Join(snapshotDir, "bots.yml"), []byte("bot: Googlebot\n"), 0o644); err != nil {
		t.Fatalf("write bots snapshot: %v", err)
	}
	if err := os.WriteFile(filepath.Join(snapshotDir, "clients.yml"), []byte("client: Firefox\n"), 0o644); err != nil {
		t.Fatalf("write clients snapshot: %v", err)
	}

	return Config{
		Version:         "v-test",
		UpstreamRepo:    "matomo-org/device-detector",
		UpstreamVersion: "v6.0.0",
		SnapshotDir:     snapshotDir,
		OutputPath:      filepath.Join(root, "out", "compiled.json"),
		ManifestPath:    filepath.Join(root, "out", "manifest.json"),
	}
}
