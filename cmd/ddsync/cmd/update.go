package ddcmd

import (
	"fmt"

	"github.com/metalagman/ddgo/internal/ddsync"
	"github.com/spf13/cobra"
)

func newUpdateCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Compile pinned snapshots into deterministic artifacts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireUpstreamVersion(opts); err != nil {
				return err
			}

			cfg := opts.config()
			manifest, err := ddsync.Update(cfg)
			if err != nil {
				return err
			}
			if err := updateProvenance(cfg, opts.provenancePath); err != nil {
				return err
			}
			payload := updateResult{
				Operation:       "update",
				SnapshotVersion: manifest.SnapshotVersion,
				UpstreamRepo:    manifest.UpstreamRepo,
				UpstreamVersion: manifest.UpstreamVersion,
				SourceFiles:     len(manifest.SourceFiles),
				OutputPath:      manifest.OutputPath,
				ManifestPath:    cfg.ManifestPath,
			}
			text := fmt.Sprintf(
				"updated snapshot %s (%s@%s) with %d files",
				manifest.SnapshotVersion,
				manifest.UpstreamRepo,
				manifest.UpstreamVersion,
				len(manifest.SourceFiles),
			)
			return writeOutput(cmd, opts.jsonOutput, payload, text)
		},
	}
}
