package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/metalagman/ddgo/internal/ddsync"
)

var errVerifyFailed = errors.New("verify failed")

func newVerifyCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify snapshots and generated artifact against manifest",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := opts.config()
			if err != nil {
				return err
			}

			report, err := ddsync.Status(cfg)
			if err != nil {
				return err
			}

			issues := append([]string(nil), report.Issues...)

			if len(issues) > 0 {
				issues = normalizeIssues(issues)
				if opts.jsonOutput {
					_ = writeJSONOutput(cmd, verifyResult{
						Operation: "verify",
						Clean:     false,
						Issues:    issues,
					})
				}
				return fmt.Errorf("%w: %s", errVerifyFailed, strings.Join(issues, "; "))
			}

			return opts.writeOutput(cmd, verifyResult{
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
