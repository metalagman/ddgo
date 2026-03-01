package ddcmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateVerifyStatusText(t *testing.T) {
	t.Parallel()

	snapshotDir, outputPath, manifestPath := testFiles(t)

	out, _, err := execCommand(t,
		"update",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
		"--version", "v-test",
	)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if !strings.Contains(out, "updated snapshot v-test with 2 files") {
		t.Fatalf("unexpected update output: %q", out)
	}

	out, _, err = execCommand(t,
		"verify",
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
		"--version", "v-test",
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
		"--version", "v-test",
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
		"--version", "v-test",
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
		"--version", "v-test",
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
		"--version", "v-test",
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
		"--version", "v-test",
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

func TestVersionAndCompletion(t *testing.T) {
	t.Parallel()

	prev := BuildVersion
	BuildVersion = "test-version"
	t.Cleanup(func() { BuildVersion = prev })

	out, _, err := execCommand(t, "version")
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

	var out bytes.Buffer
	var stderr bytes.Buffer
	root := NewRootCommand(&out, &stderr)
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
	return snapshotDir, filepath.Join(root, "compiled.json"), filepath.Join(root, "manifest.json")
}
