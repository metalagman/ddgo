package ddcmd

import (
	"io"
	"os"
	"strings"

	"github.com/metalagman/ddgo/internal/ddsync"
	"github.com/spf13/cobra"
)

const (
	defaultSnapshotDir = "sync/snapshots/v1"
	defaultOutputPath  = "sync/compiled.json"
	defaultManifest    = "sync/manifest.json"
	defaultVersion     = "v1"
)

// BuildVersion is set during builds via -ldflags.
var BuildVersion = "dev"

type rootOptions struct {
	snapshotDir string
	outputPath  string
	manifest    string
	version     string
	jsonOutput  bool
}

// Execute runs the ddsync root command with process stdio.
func Execute() error {
	return NewRootCommand(os.Stdout, os.Stderr).Execute()
}

// NewRootCommand creates the Cobra root command with subcommands.
func NewRootCommand(stdout, stderr io.Writer) *cobra.Command {
	opts := &rootOptions{}
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
	flags.StringVar(&opts.version, "version", defaultVersion, "Pinned snapshot version")
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

func (o *rootOptions) config() ddsync.Config {
	return ddsync.Config{
		Version:      strings.TrimSpace(o.version),
		SnapshotDir:  strings.TrimSpace(o.snapshotDir),
		OutputPath:   strings.TrimSpace(o.outputPath),
		ManifestPath: strings.TrimSpace(o.manifest),
	}
}
