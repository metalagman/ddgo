package ddsync

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeUpstreamTag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "v-prefixed",
			in:   "v6.5.0",
			want: "v6.5.0",
		},
		{
			name: "bare semver",
			in:   "6.5.0",
			want: "6.5.0",
		},
		{
			name: "trim whitespace",
			in:   "  6.5.0  ",
			want: "6.5.0",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := sanitizeUpstreamTag(tc.in)
			if err != nil {
				t.Fatalf("sanitizeUpstreamTag(%q) error = %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("sanitizeUpstreamTag(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSanitizeUpstreamTagRejectsUnsupportedTag(t *testing.T) {
	t.Parallel()

	_, err := sanitizeUpstreamTag("v6.5.0-rc1")
	if err == nil {
		t.Fatal("sanitizeUpstreamTag() expected error for pre-release tag")
	}
	if !strings.Contains(err.Error(), "unsupported upstream tag") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncSnapshotFromRepoURLCopiesRegexes(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	writeTestFile(t, filepath.Join(repoDir, "regexes", "bots.yml"), "bot: Googlebot\n")
	writeTestFile(t, filepath.Join(repoDir, "regexes", "device", "mobiles.yml"), "device: phone\n")
	writeTestFile(t, filepath.Join(repoDir, "README.md"), "not part of snapshot\n")
	initTaggedRepo(t, repoDir, "v1.0.0")

	snapshotDir := t.TempDir()
	writeTestFile(t, filepath.Join(snapshotDir, "stale.txt"), "stale\n")

	if err := syncSnapshotFromRepoURL(context.Background(), repoDir, "v1.0.0", snapshotDir, nil); err != nil {
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
	err := syncSnapshotFromRepoURL(context.Background(), repoDir, "v1.0.0", snapshotDir, nil)
	if err == nil {
		t.Fatal("expected error for missing regexes dir")
	}
	if !strings.Contains(err.Error(), "upstream regex directory not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncSnapshotFromRepoURLRefusesUnsafeDir(t *testing.T) {
	t.Parallel()

	err := syncSnapshotFromRepoURL(context.Background(), "repo", "v1.0.0", ".", nil)
	if err == nil || !errors.Is(err, errInvalidSnapshotDir) {
		t.Errorf("got error %v, want %v", err, errInvalidSnapshotDir)
	}
}

func TestSyncSnapshotFromUpstream(t *testing.T) {
	t.Parallel()

	cfg := Config{
		UpstreamRepo:    "matomo-org/device-detector",
		UpstreamVersion: "v6.0.0",
		SnapshotDir:     t.TempDir(),
	}

	t.Run("success", func(t *testing.T) {
		cloner := func(ctx context.Context, repoURL, tag, dstDir string) error {
			return os.WriteFile(filepath.Join(dstDir, "bots.yml"), []byte("bot: Googlebot\n"), 0o644)
		}
		err := syncSnapshotFromUpstream(context.Background(), cfg, cloner)
		if err != nil {
			t.Fatalf("syncSnapshotFromUpstream failed: %v", err)
		}
		assertFileContent(t, filepath.Join(cfg.SnapshotDir, "bots.yml"), "bot: Googlebot\n")
	})

	t.Run("invalid repo", func(t *testing.T) {
		badCfg := cfg
		badCfg.UpstreamRepo = "invalid"
		err := syncSnapshotFromUpstream(context.Background(), badCfg, nil)
		if err == nil {
			t.Fatal("expected error for invalid repo")
		}
	})

	t.Run("missing version", func(t *testing.T) {
		badCfg := cfg
		badCfg.UpstreamVersion = ""
		err := syncSnapshotFromUpstream(context.Background(), badCfg, nil)
		if !errors.Is(err, errUpstreamVersionMissing) {
			t.Errorf("got error %v, want %v", err, errUpstreamVersionMissing)
		}
	})

	t.Run("cloner error", func(t *testing.T) {
		wantErr := errors.New("clone failed")
		cloner := func(ctx context.Context, repoURL, tag, dstDir string) error {
			return wantErr
		}
		err := syncSnapshotFromUpstream(context.Background(), cfg, cloner)
		if !errors.Is(err, wantErr) {
			t.Errorf("got error %v, want %v", err, wantErr)
		}
	})
}

func TestSyncSnapshotFromUpstreamWrapper(t *testing.T) {
	if os.Getenv("NETWORK_TEST") != "1" {
		t.Skip("skipping network-dependent test; set NETWORK_TEST=1 to run")
	}

	repo := "matomo-org/device-detector"
	tag, err := ResolveLatestStableTag(repo)
	if err != nil {
		t.Fatalf("ResolveLatestStableTag failed: %v", err)
	}

	cfg := Config{
		UpstreamRepo:    repo,
		UpstreamVersion: tag,
		SnapshotDir:     t.TempDir(),
	}
	err = SyncSnapshotFromUpstream(cfg)
	if err != nil {
		t.Fatalf("SyncSnapshotFromUpstream failed: %v", err)
	}
}

func TestSyncSnapshotFromRepoURLUsesCloner(t *testing.T) {
	t.Parallel()

	snapshotDir := t.TempDir()
	repoURL := "https://github.com/matomo-org/device-detector.git"
	tag := "v6.0.0"

	t.Run("cloner called", func(t *testing.T) {
		called := false
		cloner := func(ctx context.Context, url, tag, dst string) error {
			called = true
			if url != repoURL {
				t.Errorf("got url %q, want %q", url, repoURL)
			}
			return nil
		}
		err := syncSnapshotFromRepoURL(context.Background(), repoURL, tag, snapshotDir, cloner)
		if err != nil {
			t.Fatalf("syncSnapshotFromRepoURL failed: %v", err)
		}
		if !called {
			t.Fatal("cloner was not called")
		}
	})
}

func TestParseGitHubRepoSlug(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"https://github.com/matomo-org/device-detector.git", "matomo-org/device-detector", true},
		{"https://github.com/matomo-org/device-detector", "matomo-org/device-detector", true},
		{" matomo-org/device-detector ", "", false}, // needs prefix
		{"https://github.com/other/repo", "", false},
		{"invalid", "", false},
	}

	for _, tc := range cases {
		got, err := parseGitHubRepoSlug(tc.in)
		if (err == nil) != tc.ok {
			t.Errorf("parseGitHubRepoSlug(%q) ok = %v, want %v (err=%v)", tc.in, err == nil, tc.ok, err)
		}
		if err == nil && got != tc.want {
			t.Errorf("parseGitHubRepoSlug(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSafeJoin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		base string
		rel  string
		want string
		ok   bool
	}{
		{"/tmp", "file.txt", "/tmp/file.txt", true},
		{"/tmp", "subdir/file.txt", "/tmp/subdir/file.txt", true},
		{"/tmp", "../file.txt", "", false},
		{"/tmp", "../../etc/passwd", "", false},
		{"/tmp", ".", "/tmp", true},
	}

	for _, tc := range cases {
		got, err := safeJoin(tc.base, tc.rel)
		if (err == nil) != tc.ok {
			t.Errorf("safeJoin(%q, %q) ok = %v, want %v (err=%v)", tc.base, tc.rel, err == nil, tc.ok, err)
		}
		if err == nil {
			// Compare cleaned paths
			if filepath.Clean(got) != filepath.Clean(tc.want) {
				t.Errorf("safeJoin(%q, %q) = %q, want %q", tc.base, tc.rel, got, tc.want)
			}
		}
	}
}

func TestCopyDirectoryContentsEmpty(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	err := copyDirectoryContents(src, dst)
	if err == nil || !strings.Contains(err.Error(), "upstream regex directory not found") {
		t.Errorf("copyDirectoryContents() error = %v; want error containing 'not found'", err)
	}
}

func TestCopyDirectoryContentsDiverseFiles(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	writeTestFile(t, filepath.Join(src, "regular.txt"), "data")
	if err := os.MkdirAll(filepath.Join(src, "subdir", "nested-dir"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeTestFile(t, filepath.Join(src, "subdir", "nested.txt"), "nested")
	writeTestFile(t, filepath.Join(src, "subdir", "nested-dir", "deep.txt"), "deep")

	// Create a symlink (non-regular file)
	if err := os.Symlink("/etc/passwd", filepath.Join(src, "link")); err != nil {
		t.Logf("skipping symlink test: %v", err)
	}

	err := copyDirectoryContents(src, dst)
	if err != nil {
		t.Fatalf("copyDirectoryContents failed: %v", err)
	}

	assertFileContent(t, filepath.Join(dst, "regular.txt"), "data")
	assertFileContent(t, filepath.Join(dst, "subdir", "nested.txt"), "nested")

	// Symlink should be skipped
	if _, err := os.Lstat(filepath.Join(dst, "link")); !os.IsNotExist(err) {
		t.Errorf("symlink was not skipped: %v", err)
	}
}

func TestCopyDirectoryContentsMkdirError(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := t.TempDir()

	if err := os.MkdirAll(filepath.Join(src, "subdir"), 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	writeTestFile(t, filepath.Join(src, "subdir", "file.txt"), "data")

	// Create a file in dst where subdir should be
	writeTestFile(t, filepath.Join(dst, "subdir"), "im-a-file")

	err := copyDirectoryContents(src, dst)
	if err == nil {
		t.Fatal("copyDirectoryContents() expected error for mkdir failure")
	}
}

func TestClearDirectoryError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subdir := filepath.Join(dir, "locked")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create a file that is hard to remove? Actually os.RemoveAll is very persistent.
	// But os.ReadDir(dir) will fail if we don't have permissions on dir itself.
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Skip("chmod failed")
	}
	defer os.Chmod(dir, 0o755)

	err := clearDirectory(dir)
	if err == nil {
		t.Fatal("clearDirectory() expected error")
	}
}

func TestWriteSnapshotFileError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "file")
	writeTestFile(t, filePath, "data")

	// Try to write to a path where parent is a file
	err := writeSnapshotFile(filepath.Join(filePath, "nested"), []byte("data"))
	if err == nil {
		t.Fatal("writeSnapshotFile() expected error")
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
