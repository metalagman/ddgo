package ddsync

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestLatestStableTagFromRemote(t *testing.T) {
	t.Parallel()

	output := []byte(`aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa	refs/tags/v6.0.0
bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb	refs/tags/v6.0.1-rc1
cccccccccccccccccccccccccccccccccccccccc	refs/tags/v5.9.9
dddddddddddddddddddddddddddddddddddddddd	refs/tags/v10.2.3
eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee	refs/tags/not-a-version
`)

	tag, err := latestStableTagFromRemote(output)
	if err != nil {
		t.Fatalf("latestStableTagFromRemote() error = %v", err)
	}
	if tag != "v10.2.3" {
		t.Fatalf("latestStableTagFromRemote() = %q, want %q", tag, "v10.2.3")
	}
}

func TestLatestStableTagFromRemoteNoValidTags(t *testing.T) {
	t.Parallel()

	output := []byte(`aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa	refs/tags/v6.0.1-rc1
bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb	refs/tags/release-candidate
`)

	_, err := latestStableTagFromRemote(output)
	if err == nil {
		t.Fatal("latestStableTagFromRemote() expected error when no stable tags exist")
	}
}

func TestParseStableSemver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		tag  string
		ok   bool
		want stableSemver
	}{
		{tag: "v1.2.3", ok: true, want: stableSemver{major: 1, minor: 2, patch: 3}},
		{tag: "6.5.0", ok: true, want: stableSemver{major: 6, minor: 5, patch: 0}},
		{tag: "v1.2", ok: false, want: stableSemver{}},
		{tag: "v1.2.3.4", ok: false, want: stableSemver{}},
		{tag: "v1.x.3", ok: false, want: stableSemver{}},
		{tag: "v-1.2.3", ok: false, want: stableSemver{}},
		{tag: "", ok: false, want: stableSemver{}},
		{tag: "1.2.3", ok: true, want: stableSemver{major: 1, minor: 2, patch: 3}},
		{tag: "v1.2.3-rc1", ok: false},
		{tag: "v1.2", ok: false},
		{tag: "v1.2.3+build", ok: false},
		{tag: "main", ok: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.tag, func(t *testing.T) {
			t.Parallel()
			got, ok := parseStableSemver(tc.tag)
			if ok != tc.ok {
				t.Fatalf("parseStableSemver(%q) ok = %v, want %v", tc.tag, ok, tc.ok)
			}
			if ok && got != tc.want {
				t.Fatalf("parseStableSemver(%q) = %+v, want %+v", tc.tag, got, tc.want)
			}
		})
	}
}

func TestResolveLatestStableTag(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := "matomo-org/device-detector"

	t.Run("success", func(t *testing.T) {
		lister := func(ctx context.Context, url string) ([]byte, error) {
			return []byte("hash\trefs/tags/v6.0.0\nhash\trefs/tags/v6.1.0\n"), nil
		}
		tag, err := resolveLatestStableTag(ctx, repo, lister)
		if err != nil {
			t.Fatalf("resolveLatestStableTag failed: %v", err)
		}
		if tag != "v6.1.0" {
			t.Errorf("got %q, want v6.1.0", tag)
		}
	})

	t.Run("unsupported repo", func(t *testing.T) {
		lister := func(ctx context.Context, url string) ([]byte, error) {
			return nil, nil
		}
		_, err := resolveLatestStableTag(ctx, "other/repo", lister)
		if !errors.Is(err, errUnsupportedRepo) {
			t.Errorf("got error %v, want %v", err, errUnsupportedRepo)
		}
	})

	t.Run("lister error", func(t *testing.T) {
		wantErr := errors.New("network error")
		lister := func(ctx context.Context, url string) ([]byte, error) {
			return nil, wantErr
		}
		_, err := resolveLatestStableTag(ctx, repo, lister)
		if !errors.Is(err, wantErr) {
			t.Errorf("got error %v, want %v", err, wantErr)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		lister := func(ctx context.Context, url string) ([]byte, error) {
			// Simulate a context that was already cancelled to trigger the timeout error path
			return nil, context.DeadlineExceeded
		}
		// Create a context that is already expired
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
		defer cancel()

		_, err := resolveLatestStableTag(ctx, repo, lister)
		if !errors.Is(err, errResolveTagsTimedOut) {
			t.Errorf("got error %v, want %v", err, errResolveTagsTimedOut)
		}
	})

	t.Run("no stable tags", func(t *testing.T) {
		lister := func(ctx context.Context, url string) ([]byte, error) {
			return []byte("hash\trefs/tags/v1.0.0-rc1\n"), nil
		}
		_, err := resolveLatestStableTag(ctx, repo, lister)
		if !errors.Is(err, errNoStableTags) {
			t.Errorf("got error %v, want %v", err, errNoStableTags)
		}
	})
}

func TestResolveLatestStableTagWrapper(t *testing.T) {
	// This test calls the real ResolveLatestStableTag which uses network/git.
	// We only run it if a special flag or env var is set, or just try it and skip on error.
	if os.Getenv("NETWORK_TEST") != "1" {
		t.Skip("skipping network-dependent test; set NETWORK_TEST=1 to run")
	}

	tag, err := ResolveLatestStableTag("matomo-org/device-detector")
	if err != nil {
		t.Fatalf("ResolveLatestStableTag failed: %v", err)
	}
	if tag == "" {
		t.Fatal("got empty tag")
	}
}

func TestUpstreamRepoURL(t *testing.T) {
	t.Parallel()

	got := upstreamRepoURL("matomo-org/device-detector")
	want := "https://github.com/matomo-org/device-detector.git"
	if got != want {
		t.Errorf("upstreamRepoURL() = %q, want %q", got, want)
	}
}

func TestCompareStableSemver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		a    stableSemver
		b    stableSemver
		want int
	}{
		{a: stableSemver{1, 2, 3}, b: stableSemver{1, 2, 3}, want: 0},
		{a: stableSemver{2, 0, 0}, b: stableSemver{1, 9, 9}, want: 1},
		{a: stableSemver{1, 9, 9}, b: stableSemver{2, 0, 0}, want: -1},
		{a: stableSemver{1, 3, 0}, b: stableSemver{1, 2, 9}, want: 1},
		{a: stableSemver{1, 2, 9}, b: stableSemver{1, 3, 0}, want: -1},
		{a: stableSemver{1, 2, 4}, b: stableSemver{1, 2, 3}, want: 1},
		{a: stableSemver{1, 2, 3}, b: stableSemver{1, 2, 4}, want: -1},
	}

	for _, tc := range cases {
		got := compareStableSemver(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("compareStableSemver(%+v, %+v) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestSanitizeGitHubRepoSlug(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"matomo-org/device-detector", "matomo-org/device-detector", false},
		{"matomo-org/device-detector.git", "matomo-org/device-detector", false},
		{" matomo-org/device-detector ", "matomo-org/device-detector", false},
		{"", "", true},
		{"invalid-slug", "", true},
		{"other/repo", "", true},
	}

	for _, tc := range cases {
		got, err := sanitizeGitHubRepoSlug(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("sanitizeGitHubRepoSlug(%q) err = %v, wantErr %v", tc.input, err, tc.wantErr)
		}
		if err == nil && got != tc.want {
			t.Errorf("sanitizeGitHubRepoSlug(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

