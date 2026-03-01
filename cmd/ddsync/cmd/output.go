package ddcmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/metalagman/ddgo/internal/ddsync"
	"github.com/spf13/cobra"
)

type updateResult struct {
	Operation       string `json:"operation"`
	SnapshotVersion string `json:"snapshot_version"`
	SourceFiles     int    `json:"source_files"`
	OutputPath      string `json:"output_path"`
	ManifestPath    string `json:"manifest_path"`
}

type verifyResult struct {
	Operation string   `json:"operation"`
	Clean     bool     `json:"clean"`
	Issues    []string `json:"issues,omitempty"`
}

type statusResult struct {
	Operation string   `json:"operation"`
	Clean     bool     `json:"clean"`
	Issues    []string `json:"issues,omitempty"`
}

func writeOutput(cmd *cobra.Command, jsonOutput bool, payload any, text string) error {
	if jsonOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(payload)
	}
	_, err := fmt.Fprintln(cmd.OutOrStdout(), text)
	return err
}

func statusText(report ddsync.StatusReport) string {
	if report.Clean {
		return "status: clean"
	}
	var b strings.Builder
	b.WriteString("status: dirty")
	for _, issue := range report.Issues {
		b.WriteString("\n- ")
		b.WriteString(issue)
	}
	return b.String()
}
