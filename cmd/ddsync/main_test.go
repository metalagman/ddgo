package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunUpdateVerifyStatus(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	snapshotDir := filepath.Join(root, "snapshots")
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		t.Fatalf("mkdir snapshot dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(snapshotDir, "bots.yml"), []byte("bot: Googlebot\n"), 0o644); err != nil {
		t.Fatalf("write bots snapshot: %v", err)
	}

	outputPath := filepath.Join(root, "compiled.json")
	manifestPath := filepath.Join(root, "manifest.json")

	args := []string{
		"--snapshot-dir", snapshotDir,
		"--output", outputPath,
		"--manifest", manifestPath,
		"--version", "v-test",
	}

	if err := run(append([]string{"update"}, args...)); err != nil {
		t.Fatalf("run update: %v", err)
	}
	if err := run(append([]string{"verify"}, args...)); err != nil {
		t.Fatalf("run verify: %v", err)
	}
	if err := run(append([]string{"status"}, args...)); err != nil {
		t.Fatalf("run status: %v", err)
	}
}
