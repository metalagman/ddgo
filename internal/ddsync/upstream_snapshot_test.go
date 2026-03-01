package ddsync

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncSnapshotFromRepoURLCopiesRegexes(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	writeTestFile(t, filepath.Join(repoDir, "regexes", "bots.yml"), "bot: Googlebot\n")
	writeTestFile(t, filepath.Join(repoDir, "regexes", "device", "mobiles.yml"), "device: phone\n")
	writeTestFile(t, filepath.Join(repoDir, "README.md"), "not part of snapshot\n")
	initTaggedRepo(t, repoDir, "v1.0.0")

	snapshotDir := t.TempDir()
	writeTestFile(t, filepath.Join(snapshotDir, "stale.txt"), "stale\n")

	if err := syncSnapshotFromRepoURL(repoDir, "v1.0.0", snapshotDir); err != nil {
		t.Fatalf("syncSnapshotFromRepoURL() failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(snapshotDir, "stale.txt")); !os.IsNotExist(err) {
		t.Fatalf("stale file still exists after sync: err=%v", err)
	}
	assertFileContent(t, filepath.Join(snapshotDir, "bots.yml"), "bot: Googlebot\n")
	assertFileContent(t, filepath.Join(snapshotDir, "device", "mobiles.yml"), "device: phone\n")
	if _, err := os.Stat(filepath.Join(snapshotDir, "README.md")); !os.IsNotExist(err) {
		t.Fatalf("unexpected non-regex file copied: err=%v", err)
	}
}

func TestSyncSnapshotFromRepoURLMissingRegexes(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	writeTestFile(t, filepath.Join(repoDir, "README.md"), "no regex dir\n")
	initTaggedRepo(t, repoDir, "v1.0.0")

	snapshotDir := t.TempDir()
	err := syncSnapshotFromRepoURL(repoDir, "v1.0.0", snapshotDir)
	if err == nil {
		t.Fatal("expected error for missing regexes dir")
	}
	if !strings.Contains(err.Error(), "upstream regex directory not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func initTaggedRepo(t *testing.T, dir, tag string) {
	t.Helper()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "ddsync-test")
	runGit(t, dir, "config", "user.email", "ddsync-test@example.com")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "seed")
	runGit(t, dir, "tag", tag)
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, strings.TrimSpace(string(out)))
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent for %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("unexpected content for %q:\nwant=%q\ngot=%q", path, want, string(got))
	}
}
