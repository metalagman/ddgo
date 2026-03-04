package ddsync

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const upstreamCloneTimeout = 2 * time.Minute

var (
	errUpstreamVersionMissing = errors.New("upstream version is required")
	errSnapshotDirMissing     = errors.New("snapshot dir is required")
	errInvalidSnapshotDir     = errors.New("refusing to sync into unsafe snapshot dir")
	errUpstreamRegexesMissing = errors.New("upstream regex directory not found")
	errUnsupportedRepoURL     = errors.New("unsupported upstream repo url")
	errUnsupportedTag         = errors.New("unsupported upstream tag")
	errUnsafeRelativePath     = errors.New("unsafe relative path")
)

// SyncSnapshotFromUpstream refreshes snapshot inputs from the resolved upstream version.
func SyncSnapshotFromUpstream(cfg Config) error {
	_, err := sanitizeGitHubRepoSlug(cfg.UpstreamRepo)
	if err != nil {
		return err
	}
	if cfg.UpstreamVersion == "" {
		return errUpstreamVersionMissing
	}
	if cfg.SnapshotDir == "" {
		return errSnapshotDirMissing
	}

	return syncSnapshotFromRepoURL(upstreamRepoURL(cfg.UpstreamRepo), cfg.UpstreamVersion, cfg.SnapshotDir)
}

func syncSnapshotFromRepoURL(repoURL, tag, snapshotDir string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return errUpstreamVersionMissing
	}
	snapshotDir = strings.TrimSpace(snapshotDir)
	if snapshotDir == "" {
		return errSnapshotDirMissing
	}

	cleanSnapshotDir := filepath.Clean(snapshotDir)
	if cleanSnapshotDir == "." || cleanSnapshotDir == string(filepath.Separator) {
		return fmt.Errorf("%w %q", errInvalidSnapshotDir, snapshotDir)
	}

	err := os.MkdirAll(cleanSnapshotDir, directoryPerm)
	if err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}
	err = clearDirectory(cleanSnapshotDir)
	if err != nil {
		return err
	}

	localRegexDir, ok := localRegexDirectory(repoURL)
	if ok {
		info, statErr := os.Stat(localRegexDir)
		if statErr != nil || !info.IsDir() {
			return errUpstreamRegexesMissing
		}
		return copyDirectoryContents(localRegexDir, cleanSnapshotDir)
	}

	return syncSnapshotFromUpstreamTag(repoURL, tag, cleanSnapshotDir)
}

func localRegexDirectory(repoURL string) (string, bool) {
	cleanRepoPath := filepath.Clean(strings.TrimSpace(repoURL))
	info, err := os.Stat(cleanRepoPath)
	if err != nil || !info.IsDir() {
		return "", false
	}
	return filepath.Join(cleanRepoPath, "regexes"), true
}

func syncSnapshotFromUpstreamTag(repoURL, tag, dstDir string) error {
	slug, err := parseGitHubRepoSlug(repoURL)
	if err != nil {
		return err
	}
	_ = slug

	safeTag, err := sanitizeUpstreamTag(tag)
	if err != nil {
		return err
	}

	workDir, err := os.MkdirTemp("", "ddsync-upstream-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(workDir)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), upstreamCloneTimeout)
	defer cancel()

	cloneDir := filepath.Join(workDir, "repo")
	_, err = git.PlainCloneContext(ctx, cloneDir, false, &git.CloneOptions{
		URL:           upstreamRepoURL(supportedUpstreamRepoSlug),
		Depth:         1,
		ReferenceName: plumbing.NewTagReferenceName(safeTag),
		SingleBranch:  true,
		NoCheckout:    false,
		Tags:          git.AllTags,
	})
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("clone upstream repo timed out: %w", ctx.Err())
		}
		return fmt.Errorf("clone upstream repo: %w", err)
	}

	upstreamRegexDir := filepath.Join(cloneDir, "regexes")
	info, err := os.Stat(upstreamRegexDir)
	if err != nil {
		return fmt.Errorf("%w: %w", errUpstreamRegexesMissing, err)
	}
	if !info.IsDir() {
		return errUpstreamRegexesMissing
	}

	return copyDirectoryContents(upstreamRegexDir, dstDir)
}

func parseGitHubRepoSlug(repoURL string) (string, error) {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return "", errUpstreamRepoRequired
	}
	if !strings.HasPrefix(repoURL, "https://github.com/") {
		return "", fmt.Errorf("%w: %q", errUnsupportedRepoURL, repoURL)
	}

	slug := strings.TrimPrefix(repoURL, "https://github.com/")
	slug = strings.TrimSuffix(slug, ".git")
	slug, err := sanitizeGitHubRepoSlug(slug)
	if err != nil {
		return "", fmt.Errorf("%w: %q", errUnsupportedRepoURL, repoURL)
	}
	return slug, nil
}

func sanitizeUpstreamTag(tag string) (string, error) {
	safeTag := strings.TrimSpace(tag)
	_, ok := parseStableSemver(safeTag)
	if !ok {
		return "", fmt.Errorf("%w: %q", errUnsupportedTag, tag)
	}
	return safeTag, nil
}

func clearDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read snapshot dir: %w", err)
	}
	for _, entry := range entries {
		target := filepath.Join(dir, entry.Name())
		err = os.RemoveAll(target)
		if err != nil {
			return fmt.Errorf("remove stale snapshot entry %q: %w", target, err)
		}
	}
	return nil
}

func copyDirectoryContents(srcDir, dstDir string) error {
	sourceFS := os.DirFS(srcDir)
	foundEntries := false

	err := fs.WalkDir(sourceFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		foundEntries = true

		dstPath, joinErr := safeJoin(dstDir, filepath.ToSlash(path))
		if joinErr != nil {
			return joinErr
		}
		if d.IsDir() {
			mkdirErr := os.MkdirAll(dstPath, directoryPerm)
			if mkdirErr != nil {
				return fmt.Errorf("mkdir %q: %w", dstPath, mkdirErr)
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		data, readErr := fs.ReadFile(sourceFS, path)
		if readErr != nil {
			return fmt.Errorf("read %q: %w", path, readErr)
		}
		return writeSnapshotFile(dstPath, data)
	})
	if err != nil {
		return err
	}
	if !foundEntries {
		return errUpstreamRegexesMissing
	}
	return nil
}

func writeSnapshotFile(dstPath string, data []byte) error {
	parentDir := filepath.Dir(dstPath)
	err := os.MkdirAll(parentDir, directoryPerm)
	if err != nil {
		return fmt.Errorf("mkdir parent for %q: %w", dstPath, err)
	}
	err = os.WriteFile(dstPath, data, filePerm)
	if err != nil {
		return fmt.Errorf("write %q: %w", dstPath, err)
	}
	return nil
}

func safeJoin(baseDir, relPath string) (string, error) {
	cleanRelPath := filepath.Clean(filepath.FromSlash(relPath))
	if cleanRelPath == "." {
		return filepath.Clean(baseDir), nil
	}
	if cleanRelPath == ".." || strings.HasPrefix(cleanRelPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %q", errUnsafeRelativePath, relPath)
	}
	return filepath.Join(baseDir, cleanRelPath), nil
}
