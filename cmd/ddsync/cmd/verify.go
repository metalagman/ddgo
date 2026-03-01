package ddcmd

import (
	"strings"

	"github.com/metalagman/ddgo/internal/ddsync"
	"github.com/spf13/cobra"
)

func newVerifyCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify snapshots and generated artifact against manifest",
		RunE: func(cmd *cobra.Command, _ []string) error {
			err := ddsync.Verify(opts.config())
			if err != nil {
				if opts.jsonOutput {
					issues := []string{err.Error()}
					if strings.Contains(err.Error(), "; ") {
						issues = strings.Split(err.Error(), "; ")
					}
					_ = writeOutput(cmd, true, verifyResult{
						Operation: "verify",
						Clean:     false,
						Issues:    issues,
					}, "")
				}
				return err
			}
			return writeOutput(cmd, opts.jsonOutput, verifyResult{
				Operation: "verify",
				Clean:     true,
			}, "verify: clean")
		},
	}
}
