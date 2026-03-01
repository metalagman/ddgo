package ddcmd

import (
	"fmt"

	"github.com/metalagman/ddgo/internal/ddsync"
	"github.com/spf13/cobra"
)

func newStatusCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Report whether snapshots and artifact are in sync",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := opts.config()
			report, err := ddsync.Status(cfg)
			if err != nil {
				return err
			}
			if cfg.UpstreamVersion != "" {
				issues, err := provenanceIssues(cfg, opts.provenancePath)
				if err != nil {
					report.Issues = append(report.Issues, fmt.Sprintf("provenance check failed: %v", err))
				} else {
					report.Issues = append(report.Issues, issues...)
				}
				report.Issues = normalizeIssues(report.Issues)
				report.Clean = len(report.Issues) == 0
			}
			return writeOutput(cmd, opts.jsonOutput, statusResult{
				Operation: "status",
				Clean:     report.Clean,
				Issues:    report.Issues,
			}, statusText(report))
		},
	}
}
