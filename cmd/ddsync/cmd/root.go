package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/metalagman/ddgo/internal/ddsync"
)

const (
	defaultSnapshotDir  = "sync/current"
	defaultOutputPath   = "sync/compiled.json"
	defaultManifest     = "sync/manifest.json"
	defaultUpstreamRepo = "matomo-org/device-detector"
)

var (
	errEmptyUpstreamRepo          = errors.New("--upstream-repo must not be empty")
	errUpstreamResolverNotDefined = errors.New("upstream resolver is not configured")
)

type rootOptions struct {
	snapshotDir        string
	outputPath         string
	manifest           string
	upstreamRepo       string
	version            string
	jsonOutput         bool
	resolveUpstreamTag func(string) (string, error)
	syncSnapshot       func(ddsync.Config) error
}

// Execute runs the ddsync root command with process stdio.
func Execute(version string) error {
	return NewRootCommand(os.Stdout, os.Stderr, version).Execute()
}

// NewRootCommand creates the Cobra root command with subcommands.
func NewRootCommand(stdout, stderr io.Writer, version string) *cobra.Command {
	return newRootCommandWithDependencies(
		stdout,
		stderr,
		version,
		ddsync.ResolveLatestStableTag,
		ddsync.SyncSnapshotFromUpstream,
	)
}

func newRootCommandWithDependencies(
	stdout, stderr io.Writer,
	version string,
	resolver func(string) (string, error),
	syncSnapshot func(ddsync.Config) error,
) *cobra.Command {
	opts := &rootOptions{
		version:            strings.TrimSpace(version),
		resolveUpstreamTag: resolver,
		syncSnapshot:       syncSnapshot,
	}
	if opts.version == "" {
		opts.version = "dev"
	}

	cmd := &cobra.Command{
		Use:           "ddsync",
		Short:         "Manage deterministic snapshot artifacts",
		Long:          "ddsync updates, verifies, and reports status for deterministic artifacts generated from pinned snapshots.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	flags := cmd.PersistentFlags()
	flags.StringVar(&opts.snapshotDir, "snapshot-dir", defaultSnapshotDir, "Path to pinned snapshot directory")
	flags.StringVar(&opts.outputPath, "output", defaultOutputPath, "Path to generated deterministic artifact")
	flags.StringVar(&opts.manifest, "manifest", defaultManifest, "Path to generated manifest")
	flags.StringVar(&opts.upstreamRepo, "upstream-repo", defaultUpstreamRepo, "Upstream repository slug for source data")
	flags.BoolVar(&opts.jsonOutput, "json", false, "Emit structured JSON output")

	cmd.AddCommand(
		newUpdateCommand(opts),
		newVerifyCommand(opts),
		newStatusCommand(opts),
		newVersionCommand(opts),
		newCompletionCommand(cmd),
	)

	return cmd
}

func (o *rootOptions) config() (ddsync.Config, error) {
	cfg := ddsync.Config{
		UpstreamRepo: strings.TrimSpace(o.upstreamRepo),
		SnapshotDir:  strings.TrimSpace(o.snapshotDir),
		OutputPath:   strings.TrimSpace(o.outputPath),
		ManifestPath: strings.TrimSpace(o.manifest),
	}
	if cfg.UpstreamRepo == "" {
		return ddsync.Config{}, errEmptyUpstreamRepo
	}
	if o.resolveUpstreamTag == nil {
		return ddsync.Config{}, errUpstreamResolverNotDefined
	}

	upstreamVersion, err := o.resolveUpstreamTag(cfg.UpstreamRepo)
	if err != nil {
		return ddsync.Config{}, fmt.Errorf("resolve latest upstream version for %q: %w", cfg.UpstreamRepo, err)
	}
	cfg.UpstreamVersion = strings.TrimSpace(upstreamVersion)
	return cfg, nil
}

func (o *rootOptions) writeOutput(cmd *cobra.Command, payload any, text string) error {
	if o.jsonOutput {
		return writeJSONOutput(cmd, payload)
	}
	return writeTextOutput(cmd, text)
}
