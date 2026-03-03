package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/metalagman/ddgo/internal/ddsync"
)

const testResolvedUpstreamVersion = "v6.0.0"

func TestUpdateVerifyStatusText(t *testing.T) {
	t.Parallel()

	snapshotDir, outputPath, manifestPath := testFiles(t)

	out, _, err := execCommand(t,
		"update",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
	)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if !strings.Contains(out, "updated snapshot (matomo-org/device-detector@"+testResolvedUpstreamVersion+")") {
		t.Fatalf("unexpected update output: %q", out)
	}

	out, _, err = execCommand(t,
		"verify",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
	)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !strings.Contains(out, "verify: clean") {
		t.Fatalf("unexpected verify output: %q", out)
	}

	out, _, err = execCommand(t,
		"status",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
	)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(out, "status: clean") {
		t.Fatalf("unexpected status output: %q", out)
	}
}

func TestStatusJSONDirty(t *testing.T) {
	t.Parallel()

	snapshotDir, outputPath, manifestPath := testFiles(t)
	_, _, err := execCommand(t,
		"update",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
	)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(snapshotDir, "bots.yml"), []byte("bot: altered\n"), 0o644); err != nil {
		t.Fatalf("mutate snapshot: %v", err)
	}

	out, _, err := execCommand(t,
		"status",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
		"--json",
	)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}

	var payload statusResult
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode status json: %v", err)
	}
	if payload.Clean {
		t.Fatalf("expected dirty payload, got %+v", payload)
	}
	if len(payload.Issues) == 0 {
		t.Fatalf("expected issues in payload, got %+v", payload)
	}
}

func TestVerifyJSONError(t *testing.T) {
	t.Parallel()

	snapshotDir, outputPath, manifestPath := testFiles(t)
	_, _, err := execCommand(t,
		"update",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
	)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(snapshotDir, "bots.yml"), []byte("bot: altered\n"), 0o644); err != nil {
		t.Fatalf("mutate snapshot: %v", err)
	}

	out, _, err := execCommand(t,
		"verify",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
		"--json",
	)
	if err == nil {
		t.Fatal("verify expected error for dirty state")
	}

	var payload verifyResult
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode verify json: %v", err)
	}
	if payload.Clean {
		t.Fatalf("expected clean=false in verify payload, got %+v", payload)
	}
}

func TestUpdateFailsWhenResolverFails(t *testing.T) {
	t.Parallel()

	snapshotDir, outputPath, manifestPath := testFiles(t)
	_, _, err := execCommandWithResolver(t, func(string) (string, error) {
		return "", fmt.Errorf("resolver failed")
	},
		"update",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
	)
	if err == nil {
		t.Fatal("expected error when upstream resolver fails")
	}
	if !strings.Contains(err.Error(), "resolver failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateFailsWhenSnapshotSyncFails(t *testing.T) {
	t.Parallel()

	snapshotDir, outputPath, manifestPath := testFiles(t)
	var out bytes.Buffer
	var stderr bytes.Buffer
	root := newRootCommandWithDependencies(
		&out,
		&stderr,
		testResolvedUpstreamVersion,
		func(string) (string, error) { return testResolvedUpstreamVersion, nil },
		func(ddsync.Config) error { return fmt.Errorf("snapshot sync failed") },
	)
	root.SetArgs([]string{
		"update",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
	})

	_, err := root.ExecuteC()
	if err == nil {
		t.Fatal("expected update to fail when snapshot sync fails")
	}
	if !strings.Contains(err.Error(), "snapshot sync failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateRequiresUpstreamRepo(t *testing.T) {
	t.Parallel()

	snapshotDir, outputPath, manifestPath := testFiles(t)
	_, _, err := execCommand(t,
		"update",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
		"--upstream-repo", "",
	)
	if err == nil {
		t.Fatal("expected error when upstream repo is empty")
	}
	if !strings.Contains(err.Error(), "--upstream-repo must not be empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemovedVersionFlagsRejected(t *testing.T) {
	t.Parallel()

	_, _, err := execCommand(t, "status", "--version", "v1")
	if err == nil {
		t.Fatal("expected --version to be rejected")
	}
	if !strings.Contains(err.Error(), "unknown flag: --version") {
		t.Fatalf("unexpected error for --version: %v", err)
	}

	_, _, err = execCommand(t, "status", "--upstream-version", "v6.0.0")
	if err == nil {
		t.Fatal("expected --upstream-version to be rejected")
	}
	if !strings.Contains(err.Error(), "unknown flag: --upstream-version") {
		t.Fatalf("unexpected error for --upstream-version: %v", err)
	}

	_, _, err = execCommand(t, "status", "--provenance", "compliance/provenance.json")
	if err == nil {
		t.Fatal("expected --provenance to be rejected")
	}
	if !strings.Contains(err.Error(), "unknown flag: --provenance") {
		t.Fatalf("unexpected error for --provenance: %v", err)
	}
}

func TestVersionAndCompletion(t *testing.T) {
	t.Parallel()

	out, _, err := execCommandWithVersion(t, "test-version", "version")
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if strings.TrimSpace(out) != "test-version" {
		t.Fatalf("unexpected version output %q", out)
	}

	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		shell := shell
		t.Run(shell, func(t *testing.T) {
			t.Parallel()
			out, _, err := execCommand(t, "completion", shell)
			if err != nil {
				t.Fatalf("completion %s failed: %v", shell, err)
			}
			if strings.TrimSpace(out) == "" {
				t.Fatalf("completion %s returned empty output", shell)
			}
		})
	}
}

func TestUnknownCommand(t *testing.T) {
	t.Parallel()

	_, _, err := execCommand(t, "unknown")
	if err == nil {
		t.Fatal("expected unknown command error")
	}
}

func execCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	return execCommandWithResolver(t, func(string) (string, error) {
		return testResolvedUpstreamVersion, nil
	}, args...)
}

func execCommandWithResolver(t *testing.T, resolver func(string) (string, error), args ...string) (string, string, error) {
	return execCommandWithVersionAndResolver(t, "dev", resolver, args...)
}

func execCommandWithVersion(
	t *testing.T,
	version string,
	args ...string,
) (string, string, error) {
	t.Helper()
	return execCommandWithVersionAndResolver(t, version, func(string) (string, error) {
		return testResolvedUpstreamVersion, nil
	}, args...)
}

func execCommandWithVersionAndResolver(
	t *testing.T,
	version string,
	resolver func(string) (string, error),
	args ...string,
) (string, string, error) {
	t.Helper()

	var out bytes.Buffer
	var stderr bytes.Buffer
	root := newRootCommandWithDependencies(&out, &stderr, version, resolver, func(ddsync.Config) error { return nil })
	root.SetArgs(args)
	_, err := root.ExecuteC()
	return out.String(), stderr.String(), err
}

func testFiles(t *testing.T) (string, string, string) {
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
	outputPath := filepath.Join(root, "compiled.json")
	manifestPath := filepath.Join(root, "manifest.json")
	return snapshotDir, outputPath, manifestPath
}
