package ddsync

import (
	"encoding/json"
	"errors"
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

func TestStatusErrorCases(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)

	// Missing manifest
	if err := os.Remove(cfg.ManifestPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("remove manifest: %v", err)
	}
	_, err := Status(cfg)
	if err == nil {
		t.Fatal("Status() expected error for missing manifest")
	}

	// Invalid manifest JSON
	if err := os.WriteFile(cfg.ManifestPath, []byte("invalid json"), 0o644); err != nil {
		t.Fatalf("write invalid manifest: %v", err)
	}
	_, err = Status(cfg)
	if err == nil {
		t.Fatal("Status() expected error for invalid manifest JSON")
	}

	// Output hash mismatch (tampered output)
	// Already covered by TestStatusDirtyWhenOutputTampered, but let's add missing output case
	cfg = testConfig(t) // reset
	if _, err := Update(cfg); err != nil {
		t.Fatalf("Update() failed: %v", err)
	}
	if err := os.Remove(cfg.OutputPath); err != nil {
		t.Fatalf("remove output: %v", err)
	}
	_, err = Status(cfg)
	if err == nil {
		t.Fatal("Status() expected error for missing output file")
	}
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{
			name: "missing snapshot dir",
			cfg: Config{
				UpstreamRepo: "repo",
				SnapshotDir:  "",
				OutputPath:   "out",
				ManifestPath: "manifest",
			},
			wantErr: errSnapshotDirRequired,
		},
		{
			name: "missing output path",
			cfg: Config{
				UpstreamRepo: "repo",
				SnapshotDir:  "snap",
				OutputPath:   "",
				ManifestPath: "manifest",
			},
			wantErr: errOutputPathRequired,
		},
		{
			name: "missing manifest path",
			cfg: Config{
				UpstreamRepo: "repo",
				SnapshotDir:  "snap",
				OutputPath:   "out",
				ManifestPath: "",
			},
			wantErr: errManifestPathRequired,
		},
		{
			name: "upstream version set but repo missing",
			cfg: Config{
				UpstreamRepo:    "",
				UpstreamVersion: "v1",
				SnapshotDir:     "snap",
				OutputPath:      "out",
				ManifestPath:    "manifest",
			},
			wantErr: errUpstreamRepoRequiredWithVer,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Update(tc.cfg)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("Update() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestReadSnapshotMissingDir(t *testing.T) {
	t.Parallel()

	_, _, err := readSnapshot("/non/existent/dir")
	if err == nil {
		t.Fatal("readSnapshot() expected error for missing directory")
	}
}

func TestUpdateWriteFileError(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	// Create a file where a directory should be to cause writeFile to fail
	if err := os.MkdirAll(filepath.Dir(cfg.OutputPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(cfg.OutputPath, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	// Use a path that is inside the file we just created
	cfg.OutputPath = filepath.Join(cfg.OutputPath, "inside-file")

	_, err := Update(cfg)
	if err == nil {
		t.Fatal("Update() expected error for writeFile failure")
	}
}

func TestVerifyInvalidManifest(t *testing.T) {
	t.Parallel()

	cfg := testConfig(t)
	if err := os.WriteFile(cfg.ManifestPath, []byte("invalid json"), 0o644); err != nil {
		t.Fatalf("write invalid manifest: %v", err)
	}

	err := Verify(cfg)
	if err == nil {
		t.Fatal("Verify() expected error for invalid manifest JSON")
	}
}

func TestMarshalStableJSONError(t *testing.T) {
	// Channels cannot be marshaled to JSON
	_, err := marshalStableJSON(make(chan int))
	if err != nil {
		// Expected error
		return
	}
	t.Fatal("marshalStableJSON() expected error for channel")
}

func TestWriteFileMkdirError(t *testing.T) {
	// Use a path where a file exists where a directory should be
	root := t.TempDir()
	filePath := filepath.Join(root, "file")
	if err := os.WriteFile(filePath, []byte("data"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	// Try to write to a path inside the file
	err := writeFile(filepath.Join(filePath, "subdir", "target"), []byte("data"))
	if err == nil {
		t.Fatal("writeFile() expected error for mkdir failure")
	}
}

func TestReadSnapshotWalkError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	subdir := filepath.Join(root, "subdir")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// On Linux, we can make it unreadable
	if err := os.Chmod(subdir, 0o000); err != nil {
		t.Skip("chmod 000 failed, skipping walk error test")
	}
	defer os.Chmod(subdir, 0o755) // cleanup

	_, _, err := readSnapshot(root)
	if err == nil {
		t.Fatal("readSnapshot() expected walk error")
	}
}

func testConfig(t *testing.T) Config {
	t.Helper()

	root := t.TempDir()
	snapshotDir := filepath.Join(root, "snapshots")
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatalf("mkdir snapshots: %v", err)
	}
	outDir := filepath.Join(root, "out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	if err := os.WriteFile(filepath.Join(snapshotDir, "bots.yml"), []byte("bot: Googlebot\n"), 0o644); err != nil {
		t.Fatalf("write bots snapshot: %v", err)
	}
	if err := os.WriteFile(filepath.Join(snapshotDir, "clients.yml"), []byte("client: Firefox\n"), 0o644); err != nil {
		t.Fatalf("write clients snapshot: %v", err)
	}

	return Config{
		UpstreamRepo:    "matomo-org/device-detector",
		UpstreamVersion: "v6.0.0",
		SnapshotDir:     snapshotDir,
		OutputPath:      filepath.Join(root, "out", "compiled.json"),
		ManifestPath:    filepath.Join(root, "out", "manifest.json"),
	}
}
