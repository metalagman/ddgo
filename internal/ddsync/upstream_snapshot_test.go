package ddsync

import (
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
