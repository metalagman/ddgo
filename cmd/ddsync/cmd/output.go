package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/metalagman/ddgo/internal/ddsync"
)

type updateResult struct {
	Operation       string `json:"operation"`
	UpstreamRepo    string `json:"upstream_repo"`
	UpstreamVersion string `json:"upstream_version"`
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

func writeJSONOutput(cmd *cobra.Command, payload any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func writeTextOutput(cmd *cobra.Command, text string) error {
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
