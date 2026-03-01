package ddsync

import "testing"

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
