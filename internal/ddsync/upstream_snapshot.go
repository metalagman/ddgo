package ddsync

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const upstreamCloneTimeout = 2 * time.Minute

// SyncSnapshotFromUpstream refreshes snapshot inputs from the resolved upstream version.
func SyncSnapshotFromUpstream(cfg Config) error {
	if cfg.UpstreamRepo == "" {
		return fmt.Errorf("upstream repo is required")
	}
	if cfg.UpstreamVersion == "" {
		return fmt.Errorf("upstream version is required")
	}
	if cfg.SnapshotDir == "" {
		return fmt.Errorf("snapshot dir is required")
	}

	return syncSnapshotFromRepoURL(upstreamRepoURL(cfg.UpstreamRepo), cfg.UpstreamVersion, cfg.SnapshotDir)
}

func syncSnapshotFromRepoURL(repoURL, tag, snapshotDir string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return fmt.Errorf("upstream version is required")
	}
	snapshotDir = strings.TrimSpace(snapshotDir)
	if snapshotDir == "" {
		return fmt.Errorf("snapshot dir is required")
	}

	cleanSnapshotDir := filepath.Clean(snapshotDir)
	if cleanSnapshotDir == "." || cleanSnapshotDir == string(filepath.Separator) {
		return fmt.Errorf("refusing to sync into unsafe snapshot dir %q", snapshotDir)
	}

	workDir, err := os.MkdirTemp("", "ddsync-upstream-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	cloneDir := filepath.Join(workDir, "repo")
	ctx, cancel := context.WithTimeout(context.Background(), upstreamCloneTimeout)
	defer cancel()

	clone := exec.CommandContext(
		ctx,
		"git",
		"clone",
		"--depth", "1",
		"--branch", tag,
		"--single-branch",
		repoURL,
		cloneDir,
	)
	cloneOutput, err := clone.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("git clone timed out")
	}
	if err != nil {
		msg := strings.TrimSpace(string(cloneOutput))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("git clone failed: %s", msg)
	}

	upstreamRegexDir := filepath.Join(cloneDir, "regexes")
	info, err := os.Stat(upstreamRegexDir)
	if err != nil {
		return fmt.Errorf("upstream regex directory not found: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("upstream regex path is not a directory")
	}

	if err := os.MkdirAll(cleanSnapshotDir, 0o755); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}
	if err := clearDirectory(cleanSnapshotDir); err != nil {
		return err
	}
	if err := copyDirectoryContents(upstreamRegexDir, cleanSnapshotDir); err != nil {
		return err
	}
	return nil
}

func clearDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read snapshot dir: %w", err)
	}
	for _, entry := range entries {
		target := filepath.Join(dir, entry.Name())
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("remove stale snapshot entry %q: %w", target, err)
		}
	}
	return nil
}

func copyDirectoryContents(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("resolve relative path: %w", err)
		}
		if rel == "." {
			return nil
		}

		dstPath := filepath.Join(dstDir, rel)
		if d.IsDir() {
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return fmt.Errorf("mkdir %q: %w", dstPath, err)
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		return copyFile(path, dstPath)
	})
}

func copyFile(srcPath, dstPath string) error {
	in, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open %q: %w", srcPath, err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("mkdir parent for %q: %w", dstPath, err)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create %q: %w", dstPath, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %q to %q: %w", srcPath, dstPath, err)
	}
	if err := out.Chmod(0o644); err != nil {
		return fmt.Errorf("chmod %q: %w", dstPath, err)
	}
	return nil
}
