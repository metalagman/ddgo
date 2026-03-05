package ddsync

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

const upstreamResolveTimeout = 15 * time.Second

const (
	stableSemverParts         = 3
	minRemoteRefParts         = 2
	supportedUpstreamRepoSlug = "matomo-org/device-detector"
)

var (
	errUpstreamRepoRequired = errors.New("upstream repo is required")
	errResolveTagsTimedOut  = errors.New("resolve upstream tags timed out")
	errNoStableTags         = errors.New("no stable semver tags found in upstream")
	errUnsupportedRepo      = errors.New("unsupported upstream repo")
	reGitHubRepoSlug        = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
)

type stableSemver struct {
	major int
	minor int
	patch int
}

type tagLister func(ctx context.Context, repoURL string) ([]byte, error)

// ResolveLatestStableTag returns the highest stable semver tag from upstream.
func ResolveLatestStableTag(upstreamRepo string) (string, error) {
	return resolveLatestStableTag(context.Background(), upstreamRepo, listRemoteTags)
}

func resolveLatestStableTag(ctx context.Context, upstreamRepo string, lister tagLister) (string, error) {
	repoSlug, err := sanitizeGitHubRepoSlug(upstreamRepo)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, upstreamResolveTimeout)
	defer cancel()

	output, err := lister(ctx, upstreamRepoURL(repoSlug))
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", errResolveTagsTimedOut
		}
		return "", err
	}

	return latestStableTagFromRemote(output)
}

func upstreamRepoURL(upstreamRepo string) string {
	return "https://github.com/" + strings.TrimSuffix(strings.TrimSpace(upstreamRepo), ".git") + ".git"
}

func listRemoteTags(ctx context.Context, repoURL string) ([]byte, error) {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	refs, err := remote.ListContext(ctx, &git.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list upstream refs: %w", err)
	}

	var refsOutput strings.Builder
	for _, ref := range refs {
		if !ref.Name().IsTag() {
			continue
		}
		refsOutput.WriteString(ref.Hash().String())
		refsOutput.WriteByte('\t')
		refsOutput.WriteString(ref.Name().String())
		refsOutput.WriteByte('\n')
	}

	return []byte(refsOutput.String()), nil
}

func sanitizeGitHubRepoSlug(upstreamRepo string) (string, error) {
	repoSlug := strings.TrimSuffix(strings.TrimSpace(upstreamRepo), ".git")
	if repoSlug == "" {
		return "", errUpstreamRepoRequired
	}
	if !reGitHubRepoSlug.MatchString(repoSlug) {
		return "", errUnsupportedRepo
	}
	if repoSlug != supportedUpstreamRepoSlug {
		return "", fmt.Errorf("%w: %q", errUnsupportedRepo, repoSlug)
	}
	return repoSlug, nil
}

func latestStableTagFromRemote(output []byte) (string, error) {
	var (
		bestTag string
		best    stableSemver
		found   bool
	)

	for line := range strings.SplitSeq(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < minRemoteRefParts {
			continue
		}
		ref := parts[1]
		if !strings.HasPrefix(ref, "refs/tags/") {
			continue
		}

		tag := strings.TrimPrefix(ref, "refs/tags/")
		version, ok := parseStableSemver(tag)
		if !ok {
			continue
		}
		if !found || compareStableSemver(version, best) > 0 {
			found = true
			best = version
			bestTag = tag
		}
	}

	if !found {
		return "", errNoStableTags
	}
	return bestTag, nil
}

func parseStableSemver(tag string) (stableSemver, bool) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return stableSemver{}, false
	}

	noPrefix := strings.TrimPrefix(tag, "v")
	if strings.ContainsAny(noPrefix, "-+") {
		return stableSemver{}, false
	}

	parts := strings.Split(noPrefix, ".")
	if len(parts) != stableSemverParts {
		return stableSemver{}, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil || major < 0 {
		return stableSemver{}, false
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil || minor < 0 {
		return stableSemver{}, false
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil || patch < 0 {
		return stableSemver{}, false
	}

	return stableSemver{major: major, minor: minor, patch: patch}, true
}

func compareStableSemver(a, b stableSemver) int {
	if a.major != b.major {
		if a.major > b.major {
			return 1
		}
		return -1
	}
	if a.minor != b.minor {
		if a.minor > b.minor {
			return 1
		}
		return -1
	}
	if a.patch != b.patch {
		if a.patch > b.patch {
			return 1
		}
		return -1
	}
	return 0
}
