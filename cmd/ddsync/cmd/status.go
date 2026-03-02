package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/metalagman/ddgo/internal/ddsync"
)

func newStatusCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Report whether snapshots and artifact are in sync",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := opts.config()
			if err != nil {
				return err
			}

			report, err := ddsync.Status(cfg)
			if err != nil {
				return err
			}

			issues, err := provenanceIssues(cfg, opts.provenancePath)
			if err != nil {
				report.Issues = append(report.Issues, fmt.Sprintf("provenance check failed: %v", err))
			} else {
				report.Issues = append(report.Issues, issues...)
			}
			report.Issues = normalizeIssues(report.Issues)
			report.Clean = len(report.Issues) == 0

			return opts.writeOutput(cmd, statusResult{
				Operation: "status",
				Clean:     report.Clean,
				Issues:    report.Issues,
			}, statusText(report))
		},
	}
}
