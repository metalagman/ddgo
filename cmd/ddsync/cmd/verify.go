package ddcmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/metalagman/ddgo/internal/ddsync"
	"github.com/spf13/cobra"
)

func newVerifyCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify snapshots and generated artifact against manifest",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := requireUpstreamVersion(opts); err != nil {
				return err
			}

			cfg := opts.config()
			report, err := ddsync.Status(cfg)
			if err != nil {
				return err
			}

			issues := append([]string(nil), report.Issues...)
			provenanceIssues, err := provenanceIssues(cfg, opts.provenancePath)
			if err != nil {
				issues = append(issues, fmt.Sprintf("provenance check failed: %v", err))
			} else {
				issues = append(issues, provenanceIssues...)
			}

			if len(issues) > 0 {
				issues = normalizeIssues(issues)
				if opts.jsonOutput {
					_ = writeOutput(cmd, true, verifyResult{
						Operation: "verify",
						Clean:     false,
						Issues:    issues,
					}, "")
				}
				return errors.New(strings.Join(issues, "; "))
			}

			return writeOutput(cmd, opts.jsonOutput, verifyResult{
				Operation: "verify",
				Clean:     true,
			}, "verify: clean")
		},
	}
}

func normalizeIssues(issues []string) []string {
	out := make([]string, 0, len(issues))
	seen := make(map[string]struct{}, len(issues))
	for _, issue := range issues {
		issue = strings.TrimSpace(issue)
		if issue == "" {
			continue
		}
		if _, ok := seen[issue]; ok {
			continue
		}
		seen[issue] = struct{}{}
		out = append(out, issue)
	}
	return out
}
