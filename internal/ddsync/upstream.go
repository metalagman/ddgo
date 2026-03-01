package ddsync

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const upstreamResolveTimeout = 15 * time.Second

type stableSemver struct {
	major int
	minor int
	patch int
}

// ResolveLatestStableTag returns the highest stable semver tag from upstream.
func ResolveLatestStableTag(upstreamRepo string) (string, error) {
	upstreamRepo = strings.TrimSpace(upstreamRepo)
	if upstreamRepo == "" {
		return "", fmt.Errorf("upstream repo is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), upstreamResolveTimeout)
	defer cancel()

	repoURL := "https://github.com/" + strings.TrimSuffix(upstreamRepo, ".git") + ".git"
	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--tags", "--refs", repoURL)
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("git ls-remote timed out")
	}
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git ls-remote failed: %s", msg)
	}

	tag, err := latestStableTagFromRemote(output)
	if err != nil {
		return "", err
	}
	return tag, nil
}

func latestStableTagFromRemote(output []byte) (string, error) {
	var (
		bestTag string
		best    stableSemver
		found   bool
	)

	lines := bytes.Split(output, []byte{'\n'})
	for _, rawLine := range lines {
		line := strings.TrimSpace(string(rawLine))
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
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
		return "", fmt.Errorf("no stable semver tags found in upstream")
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
	if len(parts) != 3 {
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
