package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/metalagman/ddgo/internal/ddsync"
)

var errSnapshotSyncNotConfigured = errors.New("snapshot sync is not configured")

func newUpdateCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Compile pinned snapshots into deterministic artifacts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := opts.config()
			if err != nil {
				return err
			}
			if opts.syncSnapshot == nil {
				return errSnapshotSyncNotConfigured
			}
			err = opts.syncSnapshot(cfg)
			if err != nil {
				return err
			}

			manifest, err := ddsync.Update(cfg)
			if err != nil {
				return err
			}
			payload := updateResult{
				Operation:       "update",
				UpstreamRepo:    manifest.UpstreamRepo,
				UpstreamVersion: manifest.UpstreamVersion,
				SourceFiles:     len(manifest.SourceFiles),
				OutputPath:      manifest.OutputPath,
				ManifestPath:    cfg.ManifestPath,
			}
			text := fmt.Sprintf(
				"updated snapshot (%s@%s) with %d files",
				manifest.UpstreamRepo,
				manifest.UpstreamVersion,
				len(manifest.SourceFiles),
			)
			return opts.writeOutput(cmd, payload, text)
		},
	}
}
