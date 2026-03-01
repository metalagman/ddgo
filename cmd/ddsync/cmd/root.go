package ddcmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/metalagman/ddgo/internal/ddsync"
	"github.com/spf13/cobra"
)

const (
	defaultSnapshotDir  = "sync/current"
	defaultOutputPath   = "sync/compiled.json"
	defaultManifest     = "sync/manifest.json"
	defaultProvenance   = "compliance/provenance.json"
	defaultUpstreamRepo = "matomo-org/device-detector"
)

// BuildVersion is set during builds via -ldflags.
var BuildVersion = "dev"

type rootOptions struct {
	snapshotDir        string
	outputPath         string
	manifest           string
	provenancePath     string
	upstreamRepo       string
	jsonOutput         bool
	resolveUpstreamTag func(string) (string, error)
	syncSnapshot       func(ddsync.Config) error
}

// Execute runs the ddsync root command with process stdio.
func Execute() error {
	return NewRootCommand(os.Stdout, os.Stderr).Execute()
}

// NewRootCommand creates the Cobra root command with subcommands.
func NewRootCommand(stdout, stderr io.Writer) *cobra.Command {
	return newRootCommandWithDependencies(stdout, stderr, ddsync.ResolveLatestStableTag, ddsync.SyncSnapshotFromUpstream)
}

func newRootCommandWithDependencies(
	stdout, stderr io.Writer,
	resolver func(string) (string, error),
	syncSnapshot func(ddsync.Config) error,
) *cobra.Command {
	opts := &rootOptions{}
	opts.resolveUpstreamTag = resolver
	opts.syncSnapshot = syncSnapshot

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
	flags.StringVar(&opts.provenancePath, "provenance", defaultProvenance, "Path to compliance provenance metadata")
	flags.StringVar(&opts.upstreamRepo, "upstream-repo", defaultUpstreamRepo, "Upstream repository slug for source data")
	flags.BoolVar(&opts.jsonOutput, "json", false, "Emit structured JSON output")

	cmd.AddCommand(
		newUpdateCommand(opts),
		newVerifyCommand(opts),
		newStatusCommand(opts),
		newVersionCommand(),
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
		return ddsync.Config{}, fmt.Errorf("--upstream-repo must not be empty")
	}
	if o.resolveUpstreamTag == nil {
		return ddsync.Config{}, fmt.Errorf("upstream resolver is not configured")
	}

	upstreamVersion, err := o.resolveUpstreamTag(cfg.UpstreamRepo)
	if err != nil {
		return ddsync.Config{}, fmt.Errorf("resolve latest upstream version for %q: %w", cfg.UpstreamRepo, err)
	}
	cfg.UpstreamVersion = strings.TrimSpace(upstreamVersion)
	return cfg, nil
}
