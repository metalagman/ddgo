package ddcmd

import "github.com/spf13/cobra"

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print ddsync version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			jsonOutput, _ := cmd.Flags().GetBool("json")
			return writeOutput(cmd, jsonOutput, map[string]string{"version": BuildVersion}, BuildVersion)
		},
	}
}
