package ddcmd

import (
	"github.com/metalagman/ddgo/internal/ddsync"
	"github.com/spf13/cobra"
)

func newStatusCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Report whether snapshots and artifact are in sync",
		RunE: func(cmd *cobra.Command, _ []string) error {
			report, err := ddsync.Status(opts.config())
			if err != nil {
				return err
			}
			return writeOutput(cmd, opts.jsonOutput, statusResult{
				Operation: "status",
				Clean:     report.Clean,
				Issues:    report.Issues,
			}, statusText(report))
		},
	}
}
