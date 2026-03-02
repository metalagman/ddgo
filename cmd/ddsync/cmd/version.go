package cmd

import "github.com/spf13/cobra"

func newVersionCommand(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print ddsync version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return opts.writeOutput(cmd, map[string]string{"version": opts.version}, opts.version)
		},
	}
}
